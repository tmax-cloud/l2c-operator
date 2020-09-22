package tupdb

import (
	"context"
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/status"
	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
)

var log = logf.Log.WithName("controller_tupdb")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new TupDB Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTupDB{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("tupdb-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource TupDB
	err = c.Watch(&source.Kind{Type: &tmaxv1.TupDB{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner TupDB
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tmaxv1.TupDB{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileTupDB implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileTupDB{}

// ReconcileTupDB reconciles a TupDB object
type ReconcileTupDB struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a TupDB object and makes changes based on the state read
// and what is in the TupDB.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTupDB) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling TupDB")

	// Fetch the TupDB instance
	instance := &tmaxv1.TupDB{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if len(instance.Status.Conditions) == 0 {
		instance.Status.SetDefaults()
	}
	// [TODO] TupDB Analyzer

	// [TODO] Hanging

	// [TODO] deploy TiberoDB
	if err = r.makeTargetDBReady(instance); err != nil {
		return reconcile.Result{}, nil
	}
	migratePipeline := MigratePipeline(instance)
	if err := r.createAndUpdateStatus(migratePipeline, instance, "error getting/creating pipeline"); err != nil {
		reqLogger.Error(err, "Error occurred")
		return reconcile.Result{}, err
	}

	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		return reconcile.Result{}, err
	}
	log.Info("Check Target Info", "IP", instance.Status.TargetHost, "Port", instance.Status.TargetPort)

	return reconcile.Result{}, nil
}

func (r *ReconcileTupDB) makeTargetDBReady(instance *tmaxv1.TupDB) error {
	// [TODO] Get or Create
	logger := utils.NewTupLogger(tmaxv1.TupDB{}, instance.Namespace, instance.Name)

	pvc := &corev1.PersistentVolumeClaim{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: dbResourceName(instance), Namespace: instance.Namespace}, pvc); err != nil {
		if errors.IsNotFound(err) {
			pvc, err = dbPvc(instance)
			if err := r.createAndUpdateStatus(pvc, instance, "error create PVC"); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	logger.Info("PVC Created")
	service := &corev1.Service{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: dbResourceName(instance), Namespace: instance.Namespace}, service); err != nil {
		if errors.IsNotFound(err) {
			service, err = dbService(instance)
			if err := r.createAndUpdateStatus(service, instance, "error create Service"); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	logger.Info("Service Created")

	secret := &corev1.Secret{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: dbResourceName(instance), Namespace: instance.Namespace}, service); err != nil {
		if errors.IsNotFound(err) {
			secret, err = dbSecret(instance)
			if err := r.createAndUpdateStatus(secret, instance, "error create Secret"); err != nil {
				return err
			}
			logger.Info("Secret Created")

		} else {
			return err
		}
	}
	deployment := &appsv1.Deployment{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: dbResourceName(instance), Namespace: instance.Namespace}, service); err != nil {
		if errors.IsNotFound(err) {
			deployment, err = dbDeploy(instance)
			if err := r.createAndUpdateStatus(deployment, instance, "error create Deployment"); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	logger.Info("Deployment Created")
	// [TODO] Check Status

	if err := updateTupDBSpec(instance, service); err != nil {
		logger.Error(err, "DB Update failed")
		return err
	}
	log.Info("Check Target Info in ready", "IP", instance.Status.TargetHost, "Port", instance.Status.TargetPort)

	return nil
}

func updateTupDBSpec(instance *tmaxv1.TupDB, service *corev1.Service) error {
	err := fmt.Errorf("update TupDB %s failed", instance.Name)
	switch service.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		if len(service.Status.LoadBalancer.Ingress) == 0 {
			return err
		}
		instance.Status.TargetHost = service.Status.LoadBalancer.Ingress[0].IP
	default:
		return err
	}

	if len(service.Spec.Ports) == 0 {
		return err
	}
	instance.Status.TargetPort = service.Spec.Ports[0].Port

	log.Info("Check Target Info in  function", "IP", instance.Status.TargetHost, "Port", instance.Status.TargetPort)
	if instance.Status.TargetPort == 0 || instance.Status.TargetHost == "" {
		log.Info("Error Check Target Info in  function", "IP", instance.Status.TargetHost, "Port", instance.Status.TargetPort)
		return err
	}

	return nil
}

func (r *ReconcileTupDB) createAndUpdateStatus(obj interface{}, instance *tmaxv1.TupDB, msg string) error {
	if err := utils.CheckAndCreateObject(obj, instance, r.client, r.scheme, false); err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, msg, err.Error()); err != nil {
			return err
		}
		return err
	}
	return nil
}

func (r *ReconcileTupDB) updateErrorStatus(instance *tmaxv1.TupDB, key status.ConditionType, stat corev1.ConditionStatus, reason, message string) error {
	if err := r.setCondition(instance, key, stat, reason, message); err != nil {
		return err
	}
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		return err
	}
	return nil
}

func (r *ReconcileTupDB) setCondition(instance *tmaxv1.TupDB, key status.ConditionType, stat corev1.ConditionStatus, reason, message string) error {
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
