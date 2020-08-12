package l2c

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/l2c-operator/pkg/sonarqube"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
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
	err = c.Watch(&source.Kind{Type: &tmaxv1.L2c{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
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
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileL2c) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling L2c")

	finalizer := "finalizer.l2c.tmax.io"

	// Fetch the L2c instance
	instance := &tmaxv1.L2c{}
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

	// If queued to be deleted, clean up SonarQube project
	if instance.GetDeletionTimestamp() != nil {
		if err := r.sonarQube.DeleteProject(instance.Name); err != nil {
			return reconcile.Result{}, err
		}
		controllerutil.RemoveFinalizer(instance, finalizer)
		if err := r.client.Update(context.TODO(), instance); err != nil {
			log.Error(err, "cannot remove finalizer")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// If finalizer is not set, set finalizer
	if len(instance.GetFinalizers()) == 0 {
		controllerutil.AddFinalizer(instance, finalizer)
		if err := r.client.Update(context.TODO(), instance); err != nil {
			log.Error(err, "cannot add finalizer")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// Set default Conditions
	if len(instance.Status.Conditions) == 0 {
		instance.Status.SetDefaults()
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			log.Error(err, "cannot update status")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// PipelineRun status check & do something & return
	// TODO

	// Create SonarQube Project
	if err := r.sonarQube.CreateProject(instance.Name); err != nil {
		if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectReady, metav1.ConditionFalse, "cannot create sonarqube project", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}
	// Set QualityProfiles
	if err := r.sonarQube.SetQualityProfiles(instance.Name, instance.Spec.Was.From.Type); err != nil {
		if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectReady, metav1.ConditionFalse, "cannot create sonarqube project", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}

	// Generate Pipeline
	pipeline := pipeline(instance)
	if err := controllerutil.SetControllerReference(instance, pipeline, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pipeline already exists
	found := &tektonv1.Pipeline{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pipeline.Name, Namespace: pipeline.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pipeline", "Pipeline.Namespace", pipeline.Namespace, "Pipeline.Name", pipeline.Name)
		if err := r.client.Create(context.TODO(), pipeline); err != nil {
			if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectReady, metav1.ConditionFalse, "creating pipeline failed", err.Error()); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	} else if err != nil {
		if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectReady, metav1.ConditionFalse, "error getting pipeline", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}

	// Set Project Ready!
	if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectReady, metav1.ConditionTrue, "all ready", "project is ready to run"); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileL2c) setCondition(instance *tmaxv1.L2c, key metav1.RowConditionType, status metav1.ConditionStatus, reason, message string) error {
	instance.Status.SetCondition(key, status, reason, message)
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		log.Error(err, "cannot update status")
		return err
	}
	return nil
}
