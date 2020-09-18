package tupwas

import (
	"context"
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/status"
	"github.com/tmax-cloud/l2c-operator/internal/utils"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
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

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

const (
	IngressDefaultHost = "waiting.for.ingress.ready"
)

var log = logf.Log.WithName("controller_tupwas")

// Add creates a new TupWAS Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTupWAS{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("tupwas-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource TupWAS
	err = c.Watch(&source.Kind{Type: &tmaxv1.TupWAS{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner TupWAS
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.TupWAS{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.TupWAS{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.TupWAS{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &rbacv1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.TupWAS{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.TupWAS{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.TupWAS{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.TupWAS{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &networkingv1beta1.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.TupWAS{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &tektonv1.Pipeline{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.TupWAS{},
	})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &tektonv1.PipelineRun{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.TupWAS{},
	})
	if err != nil {
		return err
	}

	tupWasReconciler, isTupWasReconciler := r.(*ReconcileTupWAS)
	if isTupWasReconciler {
		log.Info("Set ingress watcher!")
		err = c.Watch(&source.Kind{Type: &networkingv1beta1.Ingress{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(tupWasReconciler.ingressMapper),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileTupWAS implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileTupWAS{}

// ReconcileTupWAS reconciles a TupWAS object
type ReconcileTupWAS struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a TupWAS object and makes changes based on the state read
// and what is in the TupWAS.Spec
func (r *ReconcileTupWAS) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling TupWAS")

	// Fetch the TupWAS instance
	instance := &tmaxv1.TupWAS{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Set default Conditions
	if len(instance.Status.Conditions) == 0 {
		instance.Status.SetDefaults()
	}

	// Watch PipelineRun
	if err := r.watchPipelineRun(instance); err != nil {
		return reconcile.Result{}, err
	}

	// Resources
	if err := r.deployResources(instance); err != nil {
		return reconcile.Result{}, err
	}

	// Update status!
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileTupWAS) updateErrorStatus(instance *tmaxv1.TupWAS, key status.ConditionType, stat corev1.ConditionStatus, reason, message string) error {
	if err := r.setCondition(instance, key, stat, reason, message); err != nil {
		return err
	}
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		return err
	}
	return nil
}

func (r *ReconcileTupWAS) setCondition(instance *tmaxv1.TupWAS, key status.ConditionType, stat corev1.ConditionStatus, reason, message string) error {
	curCond, found := instance.Status.GetCondition(key)
	if !found {
		err := fmt.Errorf("cannot find conditions %s", string(key))
		log.Error(err, "")
		return err
	}
	if curCond.Status == stat && curCond.Reason == status.ConditionReason(reason) && curCond.Message == message {
		return nil
	}

	instance.Status.Conditions = instance.Status.SetCondition(key, stat, reason, message)

	return nil
}

func (r *ReconcileTupWAS) createAndUpdateStatus(obj interface{}, instance *tmaxv1.TupWAS, msg string) error {
	if err := utils.CheckAndCreateObject(obj, instance, r.client, r.scheme, false); err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, msg, err.Error()); err != nil {
			return err
		}
		return err
	}
	return nil
}

// To watch WAS ingress - does not have TupWAS as an owner
func (r *ReconcileTupWAS) ingressMapper(ing handler.MapObject) []reconcile.Request {
	label := ing.Meta.GetLabels()
	setTupWas := ""
	isTierWas := false
	for k, v := range label {
		if k == "tupWas" {
			setTupWas = v
		} else if k == "component" && v == "was" {
			isTierWas = true
		}
	}

	if setTupWas != "" && isTierWas {
		tupWas := &tmaxv1.TupWAS{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: setTupWas, Namespace: ing.Meta.GetNamespace()}, tupWas); err != nil {
			return []reconcile.Request{}
		}
		return []reconcile.Request{{
			NamespacedName: types.NamespacedName{
				Name:      setTupWas,
				Namespace: ing.Meta.GetNamespace(),
			},
		}}
	}

	return []reconcile.Request{}
}
