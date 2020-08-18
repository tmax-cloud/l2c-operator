package v1

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
	corev1 "k8s.io/api/core/v1"
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

	if !found {
		s.Conditions = append(s.Conditions, *cond)
	}
}

func (s *L2cStatus) SetDefaults() {
	s.SetDefaultConditions()
	// TODO
}

var conditions = []status.ConditionType{ConditionKeyProjectReady, ConditionKeyProjectRunning, ConditionKeyAnalyze, ConditionKeyDbMigrate, ConditionKeyBuild, ConditionKeyDeploy}

func (s *L2cStatus) SetDefaultConditions() {
	cond := status.Condition{
		Status: corev1.ConditionFalse,
	}
	for _, t := range conditions {
		cond.Type = t
		s.Conditions = append(s.Conditions, cond)
	}
}
