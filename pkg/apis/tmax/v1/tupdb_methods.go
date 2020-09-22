package v1

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
	"github.com/tmax-cloud/l2c-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *TupDBStatus) GetCondition(key status.ConditionType) (*status.Condition, bool) {
	for i, v := range s.Conditions {
		if v.Type == key {
			return &s.Conditions[i], true
		}
	}

	return nil, false
}

func (s *TupDBStatus) SetCondition(key status.ConditionType, stat corev1.ConditionStatus, reason, message string) []status.Condition {
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

func (s *TupDBStatus) SetDefaults() {
	s.SetDefaultConditions()
}

var tupDBConditions = []status.ConditionType{DBConditionKeyDBAnalyzing, DBConditionKeyDBMigrating, DBConditionKeyDBSucceed}

func (s *TupDBStatus) SetDefaultConditions() {
	s.Conditions = nil
	// Global Conditions
	cond := status.Condition{
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
	}
	for _, t := range tupDBConditions {
		utils.NewTupLogger(TupDB{},"unknown", "unknown").Info("Set Default", "type", t)
		cond.Type = t
		if t == DBConditionKeyDBSucceed {
			cond.Status = corev1.ConditionUnknown
			cond.Reason = "Not executed or still running"
		}
		s.Conditions = append(s.Conditions, cond)
	}
}

func (t *TupDB) GenAnalyzePipelineName() string {
	return t.Name + "-analyze"
}

func (t *TupDB) GenMigratePipelineName() string {
	return t.Name + "-migrate"
}

func (t *TupDB) GenLabels() map[string]string {
	return map[string]string{
		"tupDB":     t.Name,
		"component": "tupDB",
	}
}
