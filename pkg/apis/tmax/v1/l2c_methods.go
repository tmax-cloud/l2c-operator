package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *L2cStatus) GetCondition(key metav1.RowConditionType) (*metav1.TableRowCondition, bool) {
	for i, v := range s.Conditions {
		if v.Type == key {
			return &s.Conditions[i], true
		}
	}

	return nil, false
}

func (s *L2cStatus) SetCondition(key metav1.RowConditionType, status metav1.ConditionStatus, reason, message string) {
	cond, found := s.GetCondition(key)
	if !found {
		newCond := metav1.TableRowCondition{
			Type:    key,
			Status:  status,
			Reason:  reason,
			Message: message,
		}
		s.Conditions = append(s.Conditions, newCond)
	} else {
		cond.Status = status
		cond.Reason = reason
		cond.Message = message
	}
}

func (s *L2cStatus) SetDefaults() {
	s.SetDefaultConditions()
	// TODO
}

var conditions = []metav1.RowConditionType{ConditionKeyProjectReady, ConditionKeyProjectRunning, ConditionKeyAnalyze, ConditionKeyDbMigrate, ConditionKeyBuild, ConditionKeyDeploy}

func (s *L2cStatus) SetDefaultConditions() {
	cond := metav1.TableRowCondition{
		Status: metav1.ConditionFalse,
	}
	for _, t := range conditions {
		cond.Type = t
		s.Conditions = append(s.Conditions, cond)
	}
}
