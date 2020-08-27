package l2c

import (
	"context"

	"github.com/operator-framework/operator-sdk/pkg/status"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

var taskPhaseMap = map[string]status.ConditionType{
	string(tmaxv1.PipelineTaskNameAnalyze): tmaxv1.ConditionKeyPhaseAnalyze,
	string(tmaxv1.PipelineTaskNameMigrate): tmaxv1.ConditionKeyPhaseDbMigrate,
	string(tmaxv1.PipelineTaskNameBuild):   tmaxv1.ConditionKeyPhaseBuild,
	string(tmaxv1.PipelineTaskNameDeploy):  tmaxv1.ConditionKeyPhaseDeploy,
}

func (r *ReconcileL2c) handlePipelineRun(instance *tmaxv1.L2c) error {
	pr := &tektonv1.PipelineRun{}
	if instance.Status.PipelineRunName == "" {
		prName := instance.Name
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: prName, Namespace: instance.Namespace}, pr); err != nil {
			if !errors.IsNotFound(err) {
				return err
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
				return err
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
				return err
			}
		} else {
			if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectRunning, corev1.ConditionTrue, "L2c is now running", condition.Message); err != nil {
				return err
			}
		}

		// For each TaskRun status, update phase condition / task status for l2c
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
					return err
				}
			}
		}

		// PR succeeded
		if condition.Status == corev1.ConditionTrue && condition.Reason == string(tektonv1.PipelineRunReasonSuccessful) {
			// Succeeded condition to true
			if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectSucceeded, corev1.ConditionTrue, "", ""); err != nil {
				return err
			}
			wasSvc, err := wasService(instance)
			if err != nil {
				return err
			}
			if err := utils.CheckAndCreateObject(wasSvc, nil, r.client, r.scheme, false); err != nil {
				return err
			}

			wasIngress, err := wasIngress(instance)
			if err != nil {
				return err
			}
			if err := utils.CheckAndCreateObject(wasIngress, nil, r.client, r.scheme, false); err != nil {
				return err
			}
		} else {
			// Succeeded condition to false
			if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectSucceeded, corev1.ConditionFalse, "", ""); err != nil {
				return err
			}
		}
	} else { // PipelineRun Not found but status is not false --> Set status not running...
		if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectRunning, corev1.ConditionFalse, "", ""); err != nil {
			return err
		}

		// Check if there were any phases with reason 'Running' -> change to 'Canceled'
		for i, p := range instance.Status.Phases {
			if p.Reason == tmaxv1.ReasonPhaseRunning {
				instance.Status.Phases[i].Reason = tmaxv1.ReasonPhaseCanceled
			}
		}
	}

	return nil
}
