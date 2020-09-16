package tupwas

import (
	"context"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileTupWAS) watchPipelineRun(instance *tmaxv1.TupWAS) error {
	// Watch Analyze PipelineRun
	analyzePr := &tektonv1.PipelineRun{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.GenAnalyzePipelineName(), Namespace: instance.Namespace}, analyzePr); err != nil && !errors.IsNotFound(err) {
		return err
	} else if err != nil && errors.IsNotFound(err) {
		instance.Status.SetCondition(tmaxv1.WasConditionKeyProjectAnalyzing, corev1.ConditionFalse, "PipelineRun is not running", "")
	} else if err == nil {
		instance.Status.LastAnalyzeStartTime = analyzePr.Status.StartTime
		instance.Status.LastAnalyzeCompletionTime = analyzePr.Status.CompletionTime
		if len(analyzePr.Status.Conditions) != 0 {
			condition := analyzePr.Status.Conditions[0]
			instance.Status.LastAnalyzeResult = condition.Reason

			// Analyze Running
			status := corev1.ConditionFalse
			if analyzePr.Status.CompletionTime == nil {
				status = corev1.ConditionTrue
			}
			instance.Status.SetCondition(tmaxv1.WasConditionKeyProjectAnalyzing, status, condition.Reason, condition.Message)
		}
	}

	// Watch Build/Deploy PipelineRun
	buildPr := &tektonv1.PipelineRun{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.GenBuildDeployPipelineName(), Namespace: instance.Namespace}, buildPr); err != nil && !errors.IsNotFound(err) {
		return err
	} else if err != nil && errors.IsNotFound(err) {
		instance.Status.SetCondition(tmaxv1.WasConditionKeyProjectRunning, corev1.ConditionFalse, "PipelineRun is not running", "")
	} else if err == nil {
		instance.Status.LastBuildStartTime = buildPr.Status.StartTime
		instance.Status.LastBuildCompletionTime = buildPr.Status.CompletionTime
		if len(buildPr.Status.Conditions) != 0 {
			condition := buildPr.Status.Conditions[0]
			instance.Status.LastBuildResult = condition.Reason

			// Build/Deploy Running
			status := corev1.ConditionFalse
			if buildPr.Status.CompletionTime == nil {
				status = corev1.ConditionTrue
			}
			instance.Status.SetCondition(tmaxv1.WasConditionKeyProjectRunning, status, condition.Reason, condition.Message)

			// Build/Deploy Complete
			if condition.Reason == string(tektonv1.PipelineRunReasonSuccessful) {
				instance.Status.SetCondition(tmaxv1.WasConditionKeyProjectSucceeded, corev1.ConditionTrue, "", "")

				// Service for WAS deployment
				wasService, err := wasService(instance)
				if err != nil {
					if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating service", err.Error()); err != nil {
						return err
					}
					return err
				}
				if err := utils.CheckAndCreateObject(wasService, nil, r.client, r.scheme, false); err != nil {
					if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating service", err.Error()); err != nil {
						return err
					}
					return err
				}

				// Ingress for WAS deployment
				wasIngress, err := wasIngress(instance)
				if err != nil {
					if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating ingress", err.Error()); err != nil {
						return err
					}
					return err
				}
				if err := utils.CheckAndCreateObject(wasIngress, nil, r.client, r.scheme, false); err != nil {
					if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating ingress", err.Error()); err != nil {
						return err
					}
					return err
				}
			}
		}
	}

	return nil
}
