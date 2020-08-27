package l2c

import (
	"context"
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/status"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	"github.com/tmax-cloud/l2c-operator/pkg/sonarqube"
)

const (
	IngressDefaultHost = "waiting.for.ingress.ready"
)

var log = logf.Log.WithName("controller_l2c")

// Add creates a new L2c Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, sonar *sonarqube.SonarQube) error {
	return add(mgr, newReconciler(mgr, sonar))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, sonar *sonarqube.SonarQube) reconcile.Reconciler {
	return &ReconcileL2c{client: mgr.GetClient(), scheme: mgr.GetScheme(), sonarQube: sonar}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("l2c-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource L2c
	if err = c.Watch(&source.Kind{Type: &tmaxv1.L2c{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner L2c
	err = c.Watch(&source.Kind{Type: &tektonv1.Pipeline{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.L2c{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &tektonv1.PipelineRun{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.L2c{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &networkingv1beta1.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.L2c{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.L2c{},
	})
	if err != nil {
		return err
	}

	l2cReconciler, isL2cReconciler := r.(*ReconcileL2c)
	if isL2cReconciler {
		log.Info("Set ingress watcher!")
		err = c.Watch(&source.Kind{Type: &networkingv1beta1.Ingress{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(l2cReconciler.ingressMapper),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileL2c implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileL2c{}

// ReconcileL2c reconciles a L2c object
type ReconcileL2c struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme

	sonarQube *sonarqube.SonarQube
}

// Reconcile reads that state of the cluster for a L2c object and makes changes based on the state read
// and what is in the L2c.Spec
func (r *ReconcileL2c) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling L2c")

	// Fetch the L2c instance
	instance := &tmaxv1.L2c{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Finalizer - Set finalizer / handle if exist
	escape, err := r.handleFinalizer(instance)
	if err != nil {
		log.Error(err, "")
		return reconcile.Result{}, err
	} else if escape {
		return reconcile.Result{}, nil
	}

	// Password handler - encrypt it symmetrically
	escape, err = r.handlePassword(instance)
	if err != nil {
		log.Error(err, "")
		return reconcile.Result{}, err
	} else if escape {
		return reconcile.Result{}, nil
	}

	// !!IMPORTANT!!
	// From here, it's all about status field
	// All changes should be aggregated and updated as a whole at the end of the reconcile loop

	/*
	 * Pre-processing for L2c project
	 */
	if err := r.makeReady(instance); err != nil {
		log.Error(err, "")
		return reconcile.Result{}, err
	}

	/*
	 * PipelineRun status
	 */
	// PipelineRun status check & do something & return
	if err := r.handlePipelineRun(instance); err != nil {
		log.Error(err, "")
		return reconcile.Result{}, err
	}

	/*
	 * Analyze failure handler
	 */
	if err = r.handleAnalyzeFailure(instance); err != nil {
		log.Error(err, "")
		return reconcile.Result{}, err
	} else if escape {
		return reconcile.Result{}, nil
	}

	// Check if WAS/IDE ingress is not configured yet
	if err := r.updateWasIngressUrl(instance); err != nil {
		log.Error(err, "")
		return reconcile.Result{}, err
	}
	if err := r.updateIdeIngressUrl(instance); err != nil {
		log.Error(err, "")
		return reconcile.Result{}, err
	}

	// Update status!
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileL2c) updateErrorStatus(instance *tmaxv1.L2c, key status.ConditionType, stat corev1.ConditionStatus, reason, message string) error {
	if err := r.setCondition(instance, key, stat, reason, message); err != nil {
		return err
	}
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		return err
	}
	return nil
}

func (r *ReconcileL2c) setCondition(instance *tmaxv1.L2c, key status.ConditionType, stat corev1.ConditionStatus, reason, message string) error {
	arr, err := r.setConditionField(instance.Status.Conditions, instance, key, stat, reason, message)
	if err != nil {
		return err
	}

	instance.Status.Conditions = arr

	return nil
}

func (r *ReconcileL2c) setPhase(instance *tmaxv1.L2c, key status.ConditionType, stat corev1.ConditionStatus, reason, message string) error {
	arr, err := r.setConditionField(instance.Status.Phases, instance, key, stat, reason, message)
	if err != nil {
		return err
	}

	instance.Status.Phases = arr

	return nil
}

func (r *ReconcileL2c) setConditionField(field []status.Condition, instance *tmaxv1.L2c, key status.ConditionType, stat corev1.ConditionStatus, reason, message string) ([]status.Condition, error) {
	curCond, found := instance.Status.GetConditionField(field, key)
	if !found {
		err := fmt.Errorf("cannot find conditions %s", string(key))
		log.Error(err, "")
		return nil, err
	}
	if curCond.Status == stat && curCond.Reason == status.ConditionReason(reason) && curCond.Message == message {
		return field, nil
	}

	return instance.Status.SetConditionField(field, key, stat, reason, message), nil
}

func (r *ReconcileL2c) createAndUpdateStatus(obj interface{}, instance *tmaxv1.L2c, msg string) error {
	if err := utils.CheckAndCreateObject(obj, instance, r.client, r.scheme, false); err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, msg, err.Error()); err != nil {
			return err
		}
		return err
	}
	return nil
}

func (r *ReconcileL2c) updateWasIngressUrl(instance *tmaxv1.L2c) error {
	return r.updateIngressUrl(instance, wasResourceName(instance), "", &instance.Status.WasUrl)
}

func (r *ReconcileL2c) updateIdeIngressUrl(instance *tmaxv1.L2c) error {
	if instance.Status.Editor == nil {
		instance.Status.Editor = &tmaxv1.EditorStatus{}
	}
	return r.updateIngressUrl(instance, ideResourceName(instance), "ide.", &instance.Status.Editor.Url)
}

func (r *ReconcileL2c) updateIngressUrl(instance *tmaxv1.L2c, ingName, hostPrefix string, statusField *string) error {
	ing := &networkingv1beta1.Ingress{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: ingName, Namespace: instance.Namespace}, ing); err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil && len(ing.Status.LoadBalancer.Ingress) != 0 && len(ing.Spec.Rules) == 1 && ing.Spec.Rules[0].Host == IngressDefaultHost {
		// If Loadbalancer is given to the ingress, but host is not set, set host!
		ing.Spec.Rules[0].Host = fmt.Sprintf("%s%s.%s.%s.nip.io", hostPrefix, instance.Name, instance.Namespace, ing.Status.LoadBalancer.Ingress[0].IP)
		if err := r.client.Update(context.TODO(), ing); err != nil {
			return err
		}
	} else if len(ing.Spec.Rules) == 1 && ing.Spec.Rules[0].Host != IngressDefaultHost {
		// Update ingress url to a status field
		*statusField = fmt.Sprintf("http://%s", ing.Spec.Rules[0].Host)
	}

	return nil
}

// To watch WAS ingress - does not have l2c as an owner
func (r *ReconcileL2c) ingressMapper(ing handler.MapObject) []reconcile.Request {
	label := ing.Meta.GetLabels()
	for k, v := range label {
		if k == "l2c" {
			l2c := &tmaxv1.L2c{}
			if err := r.client.Get(context.TODO(), types.NamespacedName{Name: v, Namespace: ing.Meta.GetNamespace()}, l2c); err != nil {
				return []reconcile.Request{}
			}
			return []reconcile.Request{{
				NamespacedName: types.NamespacedName{
					Name:      v,
					Namespace: ing.Meta.GetNamespace(),
				},
			}}
		}
	}

	return []reconcile.Request{}
}
