package sonarqube

import (
	"fmt"
	"net/http"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

const (
	QualityGateName = "migration"
)

func (s *SonarQube) SetQualityGate() error {
	var gateId int32 = -1
	isGateDefault := false
	// Check if there exists QualityGate
	listResult := &tmaxv1.SonarQubeQualityGateListResult{}
	if err := s.reqHttp(http.MethodGet, "/api/qualitygates/list", nil, nil, listResult); err != nil {
		return err
	}
	for _, g := range listResult.QualityGates {
		if g.Name == QualityGateName {
			gateId = g.ID
			isGateDefault = g.IsDefault
		}
	}

	// Create QualityGate only if gate is not found
	if gateId < 0 {
		createResult := &tmaxv1.SonarQualityGateCreateResult{}
		if err := s.reqHttp(http.MethodPost, "/api/qualitygates/create", map[string]string{"name": QualityGateName}, nil, createResult); err != nil {
			return err
		}
		gateId = createResult.ID
		log.Info("Created a QualityGate")
	}

	condMetric := "violations"
	condOP := "GT"
	condError := fmt.Sprintf("%d", 0)

	// Check if condition exists
	showResult := &tmaxv1.SonarQubeQualityGateShowResult{}
	if err := s.reqHttp(http.MethodGet, "/api/qualitygates/show", map[string]string{"id": fmt.Sprintf("%d", gateId)}, nil, showResult); err != nil {
		return err
	}
	condFound := false
	for _, c := range showResult.Conditions {
		if c.Metric == condMetric && c.OP == condOP && c.Error == condError {
			condFound = true
		}
	}

	// Set condition to the quality gate only if condition is not found
	if !condFound {
		conditionData := map[string]string{
			"gateId": fmt.Sprintf("%d", gateId),
			"metric": condMetric,
			"op":     condOP,
			"error":  condError,
		}
		if err := s.reqHttp(http.MethodPost, "/api/qualitygates/create_condition", conditionData, nil, nil); err != nil {
			return err
		}
		log.Info("Created a QualityGate Condition")
	}

	// Set as default only if the gate is not a default
	if !isGateDefault {
		if err := s.reqHttp(http.MethodPost, "/api/qualitygates/set_as_default", map[string]string{"id": fmt.Sprintf("%d", gateId)}, nil, nil); err != nil {
			return err
		}
		log.Info("Set the QualityGate as default")
	}

	return nil
}
