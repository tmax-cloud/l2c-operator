package v1

import (
	"fmt"
	"strings"

	"github.com/operator-framework/operator-sdk/pkg/status"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (l *L2c) GetSonarProjectName() string {
	// Project key : <Namespace>_<Name>
	// It's valid as l2c cannot have underscore(_) for its name/namespace
	return fmt.Sprintf("%s_%s", l.Namespace, l.Name)
}

func (s *L2cStatus) GetCondition(key status.ConditionType) (*status.Condition, bool) {
	return s.GetConditionField(s.Conditions, key)
}

func (s *L2cStatus) SetCondition(key status.ConditionType, stat corev1.ConditionStatus, reason, message string) {
	s.Conditions = s.SetConditionField(s.Conditions, key, stat, reason, message)
}

func (s *L2cStatus) GetPhase(key status.ConditionType) (*status.Condition, bool) {
	return s.GetConditionField(s.Phases, key)
}

func (s *L2cStatus) SetPhase(key status.ConditionType, stat corev1.ConditionStatus, reason, message string) {
	s.Phases = s.SetConditionField(s.Phases, key, stat, reason, message)
}

func (s *L2cStatus) GetConditionField(field []status.Condition, key status.ConditionType) (*status.Condition, bool) {
	for i, v := range field {
		if v.Type == key {
			return &field[i], true
		}
	}

	return nil, false
}

func (s *L2cStatus) SetConditionField(field []status.Condition, key status.ConditionType, stat corev1.ConditionStatus, reason, message string) []status.Condition {
	cond, found := s.GetConditionField(field, key)
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
		return append(field, *cond)
	}

	return field
}

func (s *L2cStatus) SetDefaults() {
	s.SetDefaultConditions()
	s.SetDefaultPhases()
}

var conditions = []status.ConditionType{ConditionKeyProjectReady, ConditionKeyProjectRunning, ConditionKeyProjectSucceeded}

func (s *L2cStatus) SetDefaultConditions() {
	s.Conditions = nil
	// Global Conditions
	cond := status.Condition{
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
	}
	for _, t := range conditions {
		cond.Type = t
		if t == ConditionKeyProjectSucceeded {
			cond.Status = corev1.ConditionUnknown
			cond.Reason = "Not executed or still running"
		}
		s.Conditions = append(s.Conditions, cond)
	}
}

var phases = []status.ConditionType{ConditionKeyPhaseAnalyze, ConditionKeyPhaseDbMigrate, ConditionKeyPhaseBuild, ConditionKeyPhaseDeploy}

func (s *L2cStatus) SetDefaultPhases() {
	s.Phases = nil
	// L2c Phases
	phase := status.Condition{
		Status:             corev1.ConditionUnknown,
		Reason:             ReasonPhaseNotExecuted,
		LastTransitionTime: metav1.Now(),
	}
	for _, t := range phases {
		phase.Type = t
		s.Phases = append(s.Phases, phase)
	}
}

func (s *L2cStatus) SetIssues(issues []SonarIssue) {
	s.SonarIssues = nil

	for _, i := range issues {
		issue := CodeIssue{
			File:    strings.TrimPrefix(i.Component, fmt.Sprintf("%s:", i.Project)),
			Line:    i.Line,
			Message: i.Message,
		}

		s.SonarIssues = append(s.SonarIssues, issue)
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
}
