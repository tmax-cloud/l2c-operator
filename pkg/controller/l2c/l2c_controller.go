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
		instance.Status.SetDefaultPhases()
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

		// PR succeeded
		if condition.Status == corev1.ConditionTrue && condition.Reason == string(tektonv1.PipelineRunReasonSuccessful) {
			// Succeeded condition to true
			if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectSucceeded, corev1.ConditionTrue, "", ""); err != nil {
				return reconcile.Result{}, err
			}
			wasSvc, err := wasService(instance)
			if err != nil {
				log.Error(err, "")
				return reconcile.Result{}, err
			}
			if err := utils.CheckAndCreateObject(wasSvc, nil, r.client, r.scheme, false); err != nil {
				if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating Service", err.Error()); err != nil {
					return reconcile.Result{}, err
				}
				return reconcile.Result{}, err
			}

			wasIngress, err := wasIngress(instance)
			if err != nil {
				log.Error(err, "")
				return reconcile.Result{}, err
			}
			if err := utils.CheckAndCreateObject(wasIngress, nil, r.client, r.scheme, false); err != nil {
				if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating Ingress", err.Error()); err != nil {
					return reconcile.Result{}, err
				}
				return reconcile.Result{}, err
			}
		} else {
			// Succeeded condition to false
			if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectSucceeded, corev1.ConditionFalse, "", ""); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else { // PipelineRun Not found but status is not false --> Set status not running...
		if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectRunning, corev1.ConditionFalse, "", ""); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Create SonarQube Project
	if err := r.sonarQube.CreateProject(instance); err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "cannot create sonarqube project", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}
	// Set QualityProfiles
	if err := r.sonarQube.SetQualityProfiles(instance, instance.Spec.Was.From.Type); err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "cannot set sonarqube qualityProfiles", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}

	// Generate ConfigMap for WAS
	wasCm, err := wasConfigMap(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	if err := utils.CheckAndCreateObject(wasCm, instance, r.client, r.scheme, false); err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating configMap", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}

	// Generate ConfigMap/Secret for DB (only if any db configuration is set)
	if instance.Spec.Db != nil {
		dbCm, err := dbConfigMap(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		if err := utils.CheckAndCreateObject(dbCm, instance, r.client, r.scheme, false); err != nil {
			if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating configMap", err.Error()); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, err
		}

		dbSecret, err := secret(instance)
		if err != nil {
			return reconcile.Result{}, err
		}

		if err := utils.CheckAndCreateObject(dbSecret, instance, r.client, r.scheme, false); err != nil {
			if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating secret", err.Error()); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, err
		}
	}

	// Generate ServiceAccount
	sa := serviceAccount(instance)
	if err := utils.CheckAndCreateObject(sa, instance, r.client, r.scheme, false); err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating serviceAccount", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}

	// Generate RoleBinding
	rb := roleBinding(instance)
	if err := utils.CheckAndCreateObject(rb, instance, r.client, r.scheme, false); err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating roleBinding", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}

	// Generate Pipeline
	pipeline, err := pipeline(instance)
	if err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating pipeline", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}
	if err := utils.CheckAndCreateObject(pipeline, instance, r.client, r.scheme, false); err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating pipeline", err.Error()); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, err
	}

	// Set Project Ready!
	instance.Status.PipelineName = pipeline.Name
	currentReadyState, found := instance.Status.GetCondition(tmaxv1.ConditionKeyProjectReady)
	if !found {
		return reconcile.Result{}, fmt.Errorf("%s condition not found", tmaxv1.ConditionKeyProjectReady)
	}
	if currentReadyState.Status != corev1.ConditionTrue {
		if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionTrue, "Ready", "project is ready to run"); err != nil {
			return reconcile.Result{}, err
		}
	}

	// If Analyze status is Failed
	analyzeStatus, asFound := instance.Status.GetPhase(tmaxv1.ConditionKeyPhaseAnalyze)
	if asFound && analyzeStatus.Status == corev1.ConditionFalse && analyzeStatus.Reason == tmaxv1.ReasonPhaseFailed {
		// Set status.sonarIssues
		issues, err := r.sonarQube.GetIssues(instance.GetSonarProjectName())
		if err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}

		instance.Status.SetIssues(issues)

		// Generate VSCode - Secret/Service/Ingress/Deployment
		// Generate Secret
		ideSecret, err := ideSecret(instance)
		if err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}
		if err := utils.CheckAndCreateObject(ideSecret, instance, r.client, r.scheme, false); err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}
		// Check IDE Password
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: ideSecret.Name, Namespace: ideSecret.Namespace}, ideSecret); err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}
		idePassword := ideSecret.Data["password"]

		// Generate Service
		ideService, err := ideService(instance)
		if err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}
		if err := utils.CheckAndCreateObject(ideService, instance, r.client, r.scheme, false); err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}

		// Generate Ingress
		ideIngress, err := ideIngress(instance)
		if err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}
		if err := utils.CheckAndCreateObject(ideIngress, instance, r.client, r.scheme, false); err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}

		// Generate Deployment
		ideDeploy, err := ideDeployment(instance)
		if err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}
		if err := utils.CheckAndCreateObject(ideDeploy, instance, r.client, r.scheme, false); err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}

		// TODO : Status check for each objects

		// Save it to status
		if instance.Status.Editor == nil {
			instance.Status.Editor = &tmaxv1.EditorStatus{}
		}
		if instance.Status.Editor.Password != string(idePassword) {
			instance.Status.Editor.Password = string(idePassword)
		}
	} else if asFound && analyzeStatus.Status == corev1.ConditionTrue {
		instance.Status.SonarIssues = nil
	}

	// Check if WAS ingress is not configured yet
	// TODO: Refactor - reusable
	wasIng := &networkingv1beta1.Ingress{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: wasResourceName(instance), Namespace: instance.Namespace}, wasIng); err != nil && !errors.IsNotFound(err) {
		log.Error(err, "")
		return reconcile.Result{}, err
	} else if err == nil && len(wasIng.Status.LoadBalancer.Ingress) != 0 && len(wasIng.Spec.Rules) == 1 && wasIng.Spec.Rules[0].Host == IngressDefaultHost {
		// If Loadbalancer is given to the ingress, but host is not set, set host!
		wasIng.Spec.Rules[0].Host = fmt.Sprintf("%s.%s.%s.nip.io", instance.Name, instance.Namespace, wasIng.Status.LoadBalancer.Ingress[0].IP)
		if err := r.client.Update(context.TODO(), wasIng); err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}
	} else if len(wasIng.Spec.Rules) == 1 && wasIng.Spec.Rules[0].Host != IngressDefaultHost {
		instance.Status.WasUrl = fmt.Sprintf("http://%s", wasIng.Spec.Rules[0].Host)
	}

	// Check if IDE ingress is not configured yet
	// TODO: Refactor - reusable
	ideIng := &networkingv1beta1.Ingress{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: ideResourceName(instance), Namespace: instance.Namespace}, ideIng); err != nil && !errors.IsNotFound(err) {
		log.Error(err, "")
		return reconcile.Result{}, err
	} else if err == nil && len(ideIng.Status.LoadBalancer.Ingress) != 0 && len(ideIng.Spec.Rules) == 1 && ideIng.Spec.Rules[0].Host == IngressDefaultHost {
		// If Loadbalancer is given to the ingress, but host is not set, set host!
		ideIng.Spec.Rules[0].Host = fmt.Sprintf("ide.%s.%s.%s.nip.io", instance.Name, instance.Namespace, ideIng.Status.LoadBalancer.Ingress[0].IP)
		if err := r.client.Update(context.TODO(), ideIng); err != nil {
			log.Error(err, "")
			return reconcile.Result{}, err
		}
	} else if len(ideIng.Spec.Rules) == 1 && ideIng.Spec.Rules[0].Host != IngressDefaultHost {
		instance.Status.Editor.Url = fmt.Sprintf("http://%s", ideIng.Spec.Rules[0].Host)
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

// To watch WAS ingress - does not have l2c as an owner
func (r *ReconcileL2c) ingressMapper(ing handler.MapObject) []reconcile.Request {
	label := ing.Meta.GetLabels()
	for k, v := range label {
		if k == "l2c" {
			l2c := &tmaxv1.L2c{}
			if err := r.client.Get(context.TODO(), types.NamespacedName{Name: v, Namespace: ing.Meta.GetNamespace()}, l2c); err != nil {
				log.Error(err, "")
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
