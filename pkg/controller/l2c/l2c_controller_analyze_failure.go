package l2c

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

func (r *ReconcileL2c) handleAnalyzeFailure(instance *tmaxv1.L2c) error {
	analyzeStatus, asFound := instance.Status.GetPhase(tmaxv1.ConditionKeyPhaseAnalyze)
	if asFound && analyzeStatus.Status == corev1.ConditionFalse && analyzeStatus.Reason == tmaxv1.ReasonPhaseFailed {
		// Set status.sonarIssues
		issues, err := r.sonarQube.GetIssues(instance.GetSonarProjectName())
		if err != nil {
			return err
		}

		instance.Status.SetIssues(issues)

		// Generate VSCode - Secret/Service/Ingress/Deployment
		// Generate Secret
		ideSecret := ideSecret(instance, r.sonarQube)
		if err := utils.CheckAndCreateObject(ideSecret, instance, r.client, r.scheme, false); err != nil {
			return err
		}
		// Check IDE Password
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: ideSecret.Name, Namespace: ideSecret.Namespace}, ideSecret); err != nil {
			return err
		}
		idePassword := ideSecret.Data["password"]

		// Generate Service
		ideService, err := ideService(instance)
		if err != nil {
			return err
		}
		if err := utils.CheckAndCreateObject(ideService, instance, r.client, r.scheme, false); err != nil {
			return err
		}

		// Generate Ingress
		ideIngress, err := ideIngress(instance)
		if err != nil {
			return err
		}
		if err := utils.CheckAndCreateObject(ideIngress, instance, r.client, r.scheme, false); err != nil {
			return err
		}

		// Generate Deployment
		ideDeploy, err := ideDeployment(instance)
		if err != nil {
			return err
		}
		if err := utils.CheckAndCreateObject(ideDeploy, instance, r.client, r.scheme, false); err != nil {
			return err
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

	return nil
}
