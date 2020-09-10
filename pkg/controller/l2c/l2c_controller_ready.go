package l2c

import (
	"fmt"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
)

func (r *ReconcileL2c) makeReady(instance *tmaxv1.L2c) error {
	// Set default Conditions
	if len(instance.Status.Conditions) == 0 {
		instance.Status.SetDefaults()
	}

	// Generate ConfigMap for WAS
	wasCm, err := wasConfigMap(instance)
	if err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating configMap", err.Error()); err != nil {
			return err
		}
		return err
	}
	if err := r.createAndUpdateStatus(wasCm, instance, "error getting/creating configMap"); err != nil {
		return err
	}

	// Generate ServiceAccount
	sa := serviceAccount(instance)
	if err := r.createAndUpdateStatus(sa, instance, "error getting/creating serviceAccount"); err != nil {
		return err
	}

	// Generate RoleBinding
	rb := roleBinding(instance)
	if err := r.createAndUpdateStatus(rb, instance, "error getting/creating roleBinding"); err != nil {
		return err
	}

	// Generate Pipeline
	pipeline, err := pipeline(instance)
	if err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating pipeline", err.Error()); err != nil {
			return err
		}
		return err
	}
	if err := r.createAndUpdateStatus(pipeline, instance, "error getting/creating pipeline"); err != nil {
		return err
	}

	// Generate ConfigMap/Secret for DB (only if any db configuration is set)
	if instance.Spec.Db != nil {
		dbCm, err := dbConfigMap(instance)
		if err != nil {
			if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating configMap", err.Error()); err != nil {
				return err
			}
			return err
		}
		if err := r.createAndUpdateStatus(dbCm, instance, "error getting/creating configMap"); err != nil {
			return err
		}

		dbSecret, err := secret(instance)
		if err != nil {
			if err := r.updateErrorStatus(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating secret", err.Error()); err != nil {
				return err
			}
			return err
		}
		if err := r.createAndUpdateStatus(dbSecret, instance, "error getting/creating secret"); err != nil {
			return err
		}
	}

	// Set Project Ready!
	instance.Status.PipelineName = pipeline.Name
	currentReadyState, found := instance.Status.GetCondition(tmaxv1.ConditionKeyProjectReady)
	if !found {
		return fmt.Errorf("%s condition not found", tmaxv1.ConditionKeyProjectReady)
	}
	if currentReadyState.Status != corev1.ConditionTrue {
		if err := r.setCondition(instance, tmaxv1.ConditionKeyProjectReady, corev1.ConditionTrue, "Ready", "project is ready to run"); err != nil {
			return err
		}
	}

	return nil
}
