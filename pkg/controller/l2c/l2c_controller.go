package l2c

import (
	"context"
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/status"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
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

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	"github.com/tmax-cloud/l2c-operator/pkg/sonarqube"
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
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
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
		if err := r.sonarQube.DeleteProject(instance); err != nil {
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

	// !!IMPORTANT!!
	// From here, it's all about status field
	// All changes should be aggregated and updated as a whole at the end of the reconcile loop

	// Set default Conditions
	if len(instance.Status.Conditions) == 0 {
		instance.Status.SetDefaults()
	}

	// PipelineRun status check & do something & return
	pr := &tektonv1.PipelineRun{}
	if instance.Status.PipelineRunName == "" {
		prName := instance.Name
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: prName, Namespace: instance.Namespace}, pr); err != nil {
			if !errors.IsNotFound(err) {
				log.Error(err, "cannot get pipelineRun")
				return reconcile.Result{}, err
			}
		} else {
			instance.Status.PipelineRunName = prName
		}
	} else {
		prName := instance.Status.PipelineRunName
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: prName, Namespace: instance.Namespace}, pr); err != nil {
			if errors.IsNotFound(err) {
				instance.Status.PipelineRunName = ""
			} else {
				log.Error(err, "cannot get pipelineRun")
				return reconcile.Result{}, err
			}
		}
	}

	// If PipelineRun exists, begin status check!
	if pr.ResourceVersion != "" && len(pr.Status.Conditions) == 1 {
		condition := pr.Status.Conditions[0]

		// Update L2c Running status True or false, depending on the status
		if pr.Status.CompletionTime != nil {
			instance.Status.CompletionTime = pr.Status.CompletionTime
			if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectRunning, corev1.ConditionFalse, condition.Reason, condition.Message); err != nil {
				return reconcile.Result{}, err
			}
		} else {
			if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectRunning, corev1.ConditionTrue, "L2c is now running", condition.Message); err != nil {
				return reconcile.Result{}, err
			}
		}

		// For each TaskRun status, update phase condition / task status for l2c
		taskPhaseMap := map[string]status.ConditionType{
			string(tmaxv1.PipelineTaskNameAnalyze): tmaxv1.ConditionKeyPhaseAnalyze,
			string(tmaxv1.PipelineTaskNameMigrate): tmaxv1.ConditionKeyPhaseDbMigrate,
			string(tmaxv1.PipelineTaskNameBuild):   tmaxv1.ConditionKeyPhaseBuild,
			string(tmaxv1.PipelineTaskNameDeploy):  tmaxv1.ConditionKeyPhaseDeploy,
		}
		// Clear first
		instance.Status.TaskStatus = nil
		for k, v := range pr.Status.TaskRuns {
			// Update task status
			stat := tmaxv1.L2cTaskStatus{TaskRunName: k}
			stat.CopyFromTaskRunStatus(v)
			instance.Status.TaskStatus = append(instance.Status.TaskStatus, stat)

			// Update phase conditions
			phase, isKnown := taskPhaseMap[v.PipelineTaskName]
			if isKnown && len(v.Status.Conditions) == 1 {
				cond := v.Status.Conditions[0]
				if err := r.setPhase(instance, phase, cond.Status, cond.Reason, cond.Message); err != nil {
					return reconcile.Result{}, err
				}
			}
		}
	} else { // PipelineRun Not found but status is not false --> Set status not running...
		if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectRunning, corev1.ConditionFalse, "", ""); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Create SonarQube Project
	if err := r.sonarQube.CreateProject(instance); err != nil {
		if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "cannot create sonarqube project", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
	}
	// Set QualityProfiles
	if err := r.sonarQube.SetQualityProfiles(instance, instance.Spec.Was.From.Type); err != nil {
		if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "cannot create sonarqube project", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Generate ConfigMap (only if any db configuration is set)
	if instance.Spec.Db != nil {
		cm, err := configMap(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		if err := utils.CheckAndCreateObject(cm, instance, r.client, r.scheme); err != nil {
			if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating configMap", err.Error()); err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	// Generate Pipeline
	pipeline := pipeline(instance)
	if err := utils.CheckAndCreateObject(pipeline, instance, r.client, r.scheme); err != nil {
		if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating pipeline", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Set Project Ready!
	instance.Status.PipelineName = pipeline.Name
	currentReadyState, found := instance.Status.GetCondition(tmaxv1.ConditionKeyProjectReady)
	if !found {
		return reconcile.Result{}, fmt.Errorf("%s condition not found", tmaxv1.ConditionKeyProjectReady)
	}
	if currentReadyState.Status != corev1.ConditionTrue {
		if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionTrue, "all ready", "project is ready to run"); err != nil {
			return reconcile.Result{}, err
		}
	}

	// If Analyze status is Failed
	analyzeStatus, asFound := instance.Status.GetPhase(tmaxv1.ConditionKeyPhaseAnalyze)
	if asFound && analyzeStatus.Status == corev1.ConditionFalse && analyzeStatus.Reason == tmaxv1.ReasonPhaseFailed {
		// Set status.sonarIssues
		issues, err := r.sonarQube.GetIssues(sonarqube.GetSonarProjectName(instance))
		if err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}

		instance.Status.SetIssues(issues)

		// Generate VSCode
		// TODO
	} else if asFound && analyzeStatus.Status == corev1.ConditionTrue {
		instance.Status.SonarIssues = nil
	}

	// Update status!
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
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
