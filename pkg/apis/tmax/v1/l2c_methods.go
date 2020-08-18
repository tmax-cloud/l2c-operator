package v1

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *L2cStatus) GetCondition(key status.ConditionType) (*status.Condition, bool) {
	for i, v := range s.Conditions {
		if v.Type == key {
			return &s.Conditions[i], true
		}
	}

	return nil, false
}

func (s *L2cStatus) SetCondition(key status.ConditionType, stat corev1.ConditionStatus, reason, message string) {
	cond, found := s.GetCondition(key)
	if !found {
		cond = &status.Condition{
			Type: key,
		}
	}

	cond.Status = stat
	cond.Reason = status.ConditionReason(reason)
	cond.Message = message
	cond.LastTransitionTime = metav1.Now()

	if !found {
		s.Conditions = append(s.Conditions, *cond)
	}
}

func (s *L2cStatus) SetDefaults() {
	s.SetDefaultConditions()
	// TODO
}

var conditions = []status.ConditionType{ConditionKeyProjectReady, ConditionKeyProjectRunning}
var phases = []status.ConditionType{ConditionKeyPhaseAnalyze, ConditionKeyPhaseDbMigrate, ConditionKeyPhaseBuild, ConditionKeyPhaseDeploy}

func (s *L2cStatus) SetDefaultConditions() {
	// Global Conditions
	cond := status.Condition{
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
	}
	for _, t := range conditions {
		cond.Type = t
		s.Conditions = append(s.Conditions, cond)
	}

	// L2c Phases
	for _, t := range phases {
		cond.Type = t
		cond.Status = corev1.ConditionUnknown
		cond.Reason = ReasonPhaseNotExecuted
		s.Conditions = append(s.Conditions, cond)
	}
}

func (s *L2cTaskStatus) CopyFromTaskRunStatus(trStatus *tektonv1.PipelineRunTaskRunStatus) {
	// Conditions
	for _, cond := range trStatus.Status.Conditions {
		s.Conditions = append(s.Conditions, status.Condition{
			Type:               status.ConditionType(cond.Type),
			Status:             cond.Status,
			Reason:             status.ConditionReason(cond.Reason),
			Message:            cond.Message,
			LastTransitionTime: cond.LastTransitionTime.Inner,
		})
	}

	// PodName
	s.PodName = trStatus.Status.PodName

	// StartTime
	s.StartTime = trStatus.Status.StartTime

	// CompletionTime
	s.CompletionTime = trStatus.Status.CompletionTime

	// Steps
	s.Steps = append(s.Steps, trStatus.Status.Steps...)

	// Sidecars
	s.Sidecars = append(s.Sidecars, trStatus.Status.Sidecars...)

	// TaskSpec
	s.TaskSpec = trStatus.Status.TaskSpec
}
