package tupdb

import (
	"context"
	"github.com/operator-framework/operator-sdk/pkg/status"
	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
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
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
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
	// [TODO] TupDB Analyzer

	// [TODO] Hanging

	// [TODO] deploy TiberoDB
	if err = r.makeTargetDBReady(instance); err != nil {
		return reconcile.Result{}, nil
	}

	ingress, err := dbIngress(instance)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ingress.Name, Namespace: ingress.Namespace}, ingress)
	if err != nil && !errors.IsNotFound(err) {
		reqLogger.Error(err, "There is no ingress yet")
		return reconcile.Result{Requeue: true}, nil
	}
	if checkIngressAndUpdate(ingress) {
		reqLogger.Info("Ingress will be updated")
		if err := r.client.Update(context.TODO(), ingress); err != nil {
			return reconcile.Result{}, nil
		}
	} else {
		reqLogger.Info("Ingress is not well created")
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileTupDB) makeTargetDBReady(instance *tmaxv1.TupDB) error {
	// [TODO] Refactoring
	logger := utils.NewTupLogger(tmaxv1.TupDB{}, instance.Namespace, instance.Name)
	pvc, err := dbPvc(instance)
	if err != nil {
		return err
	}
	if err = r.createAndUpdateStatus(pvc, instance, "error create PVC"); err != nil {
		return err
	}
	logger.Info("PVC Created")

	service, err := dbService(instance)
	if err != nil {
		return err
	}
	if err = r.createAndUpdateStatus(service, instance, "error create Service"); err != nil {
		return err
	}
	logger.Info("Service Created")

	secret, err := dbSecret(instance)
	if err != nil {
		logger.Error(err, "Secret err")
		return err
	}
	if err = r.createAndUpdateStatus(secret, instance, "error create Secret"); err != nil {
		return err
	}
	logger.Info("Secret Created")

	deployment, err := dbDeploy(instance)
	if err != nil {
		return err
	}
	if err = r.createAndUpdateStatus(deployment, instance, "error create Deployment"); err != nil {
		return err
	}
	logger.Info("Deployment Created")

	ingress, err := dbIngress(instance)
	if err != nil {
		return err
	}
	if err = r.createAndUpdateStatus(ingress, instance, "error create Ingress"); err != nil {
		return err
	}
	logger.Info("Ingress Created")

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
	//if err := r.setCondition(instance, key, stat, reason, message); err != nil {
	//	return err
	//}
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		return err
	}
	return nil
}

//func (r *ReconcileTupDB) setCondition(instance *tmaxv1.TupDB, key status.ConditionType, stat corev1.ConditionStatus, reason, message string) error {
//	arr, err := r.setConditionField(instance.Status.Conditions, instance, key, stat, reason, message)
//	if err != nil {
//		return err
//	}
//
//	instance.Status.Conditions = arr
//
//	return nil
//}
//
//func (r *ReconcileTupDB) setConditionField(field []status.Condition, instance *tmaxv1.TupDB, key status.ConditionType, stat corev1.ConditionStatus, reason, message string) ([]status.Condition, error) {
//	curCond, found := instance.Status.GetConditionField(field, key)
//	if !found {
//		err := fmt.Errorf("cannot find conditions %s", string(key))
//		log.Error(err, "")
//		return nil, err
//	}
//	if curCond.Status == stat && curCond.Reason == status.ConditionReason(reason) && curCond.Message == message {
//		return field, nil
//	}
//
//	return instance.Status.SetConditionField(field, key, stat, reason, message), nil
//}
