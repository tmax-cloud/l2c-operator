package tupwas

import (
	"context"
	"fmt"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileTupWAS) deployResources(instance *tmaxv1.TupWAS) error {
	// Set Project Ready first
	currentReadyState, found := instance.Status.GetCondition(tmaxv1.WasConditionKeyProjectReady)
	if !found {
		return fmt.Errorf("%s condition not found", tmaxv1.WasConditionKeyProjectReady)
	}
	if currentReadyState.Status != corev1.ConditionTrue {
		if err := r.setCondition(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionTrue, "Ready", "project is ready to run"); err != nil {
			return err
		}
	}

	// PVC for git & repo
	pvc, err := gitReportPVC(instance)
	if err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating PVC", err.Error()); err != nil {
			return err
		}
		return err
	}
	if err := r.createAndUpdateStatus(pvc, instance, "error getting/creating PVC"); err != nil {
		return err
	}

	// ConfigMap for WAS deployment
	wasConfigMap, err := wasDeployConfigMap(instance)
	if err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating configMap", err.Error()); err != nil {
			return err
		}
		return err
	}
	if err := r.createAndUpdateStatus(wasConfigMap, instance, "error getting/creating configMap"); err != nil {
		return err
	}

	// ServiceAccount for WAS deployment
	wasDeploySa := wasDeployServiceAccount(instance)
	if err := r.createAndUpdateStatus(wasDeploySa, instance, "error getting/creating serviceAccount"); err != nil {
		return err
	}

	// RoleBinding for WAS deployment
	wasDeployRb := wasDeployRoleBinding(instance)
	if err := r.createAndUpdateStatus(wasDeployRb, instance, "error getting/creating roleBinding"); err != nil {
		return err
	}

	// Pipeline 1 - Analyze
	analyzePipeline := analyzePipeline(instance)
	if err := r.createAndUpdateStatus(analyzePipeline, instance, "error getting/creating pipeline"); err != nil {
		return err
	}

	// Pipeline 2 - Build/Deploy
	buildDeployPipeline, err := buildDeployPipeline(instance)
	if err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating pipeline", err.Error()); err != nil {
			return err
		}
		return err
	}
	if err := r.createAndUpdateStatus(buildDeployPipeline, instance, "error getting/creating pipeline"); err != nil {
		return err
	}

	// IDE resources
	if err := r.deployIdeReport(instance); err != nil {
		return err
	}

	// Service for WAS deployment - TODO: Only if PipelineRun complete!
	wasService, err := wasService(instance)
	if err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating service", err.Error()); err != nil {
			return err
		}
		return err
	}
	if err := r.createAndUpdateStatus(wasService, instance, "error getting/creating service"); err != nil {
		return err
	}

	// Ingress for WAS deployment - TODO: Only if PipelineRun complete!
	wasIngress, err := wasIngress(instance)
	if err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating ingress", err.Error()); err != nil {
			return err
		}
		return err
	}
	if err := r.createAndUpdateStatus(wasIngress, instance, "error getting/creating ingress"); err != nil {
		return err
	}

	// Update ingress - apply host
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: wasIngress.Name, Namespace: wasIngress.Namespace}, wasIngress)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil {
		if len(wasIngress.Status.LoadBalancer.Ingress) != 0 && len(wasIngress.Spec.Rules) == 1 && wasIngress.Spec.Rules[0].Host == IngressDefaultHost {
			// If Loadbalancer is given to the ingress, but host is not set, set host!
			wasIngress.Spec.Rules[0].Host = fmt.Sprintf("%s.%s.%s.nip.io", instance.Name, instance.Namespace, wasIngress.Status.LoadBalancer.Ingress[0].IP)
			if err := r.client.Update(context.TODO(), wasIngress); err != nil {
				return err
			}
		} else if len(wasIngress.Spec.Rules) == 1 && wasIngress.Spec.Rules[0].Host != IngressDefaultHost {
			// Update ingress url to a status field
			instance.Status.WasUrl = fmt.Sprintf("http://%s", wasIngress.Spec.Rules[0].Host)
		}
	}

	return nil
}
