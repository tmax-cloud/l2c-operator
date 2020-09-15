package v1

import (
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/status"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/l2c-operator/internal"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *TupWASStatus) GetCondition(key status.ConditionType) (*status.Condition, bool) {
	for i, v := range s.Conditions {
		if v.Type == key {
			return &s.Conditions[i], true
		}
	}

	return nil, false
}

func (s *TupWASStatus) SetCondition(key status.ConditionType, stat corev1.ConditionStatus, reason, message string) []status.Condition {
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
		return append(s.Conditions, *cond)
	}

	return s.Conditions
}

func (s *TupWASStatus) SetDefaults() {
	s.SetDefaultConditions()
}

var tupWasConditions = []status.ConditionType{WasConditionKeyProjectReady, WasConditionKeyProjectAnalyzing, WasConditionKeyProjectRunning, WasConditionKeyProjectSucceeded}

func (s *TupWASStatus) SetDefaultConditions() {
	s.Conditions = nil
	// Global Conditions
	cond := status.Condition{
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
	}
	for _, t := range tupWasConditions {
		cond.Type = t
		if t == ConditionKeyProjectSucceeded {
			cond.Status = corev1.ConditionUnknown
			cond.Reason = "Not executed or still running"
		}
		s.Conditions = append(s.Conditions, cond)
	}
}

func (s *TupWasTaskStatus) CopyFromTaskRunStatus(trStatus *tektonv1.PipelineRunTaskRunStatus) {
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
}

// Supporting functions
func (t *TupWAS) GenResourceName() string {
	return t.Name
}

func (t *TupWAS) GenLabels() map[string]string {
	return map[string]string{
		"tupWas":    t.Name,
		"component": "tupWas",
	}
}

func (t *TupWAS) GenBuilderImage() (string, error) {
	switch t.Spec.To.Type {
	case "jeus":
		return internal.BuilderImageJeus, nil
	default:
		return "", fmt.Errorf("%s was type is not supported", t.Spec.To.Type)
	}
}

func (t *TupWAS) GenAnalyzePipelineName() string {
	return t.GenResourceName() + "-analyze"
}

func (t *TupWAS) GenBuildDeployPipelineName() string {
	return t.GenResourceName() + "-build-deploy"
}

// Supporting functions for WAS resources
func (t *TupWAS) GenWasResourceName() string {
	return fmt.Sprintf("%s-was", t.Name)
}

func (t *TupWAS) GenWasLabels() map[string]string {
	return map[string]string{
		"tupWas":    t.Name,
		"component": "was",
	}
}

func (t *TupWAS) GenWasServiceLabels() map[string]string {
	return map[string]string{
		"tupWas": t.Name,
		"tier":   t.Spec.To.Type,
	}
}

func (t *TupWAS) GenWasPort() (int32, error) {
	switch t.Spec.To.Type {
	case "jeus":
		return 8808, nil
	default:
		return 0, fmt.Errorf("spec.was.to.type(%s) not supported", t.Spec.To.Type)
	}
}
