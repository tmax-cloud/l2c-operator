package l2c

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

const (
	Finalizer = "finalizer.l2c.tmax.io"
)

func (r *ReconcileL2c) handleFinalizer(instance *tmaxv1.L2c) (bool, error) {
	// If queued to be deleted, clean up SonarQube project
	if instance.GetDeletionTimestamp() != nil {
		if err := r.sonarQube.DeleteProject(instance); err != nil {
			return false, err
		}
		controllerutil.RemoveFinalizer(instance, Finalizer)
		if err := r.client.Update(context.TODO(), instance); err != nil {
			log.Error(err, "cannot remove finalizer")
			return false, err
		}
		return true, nil
	}

	// If finalizer is not set, set finalizer
	if len(instance.GetFinalizers()) == 0 {
		controllerutil.AddFinalizer(instance, Finalizer)
		if err := r.client.Update(context.TODO(), instance); err != nil {
			log.Error(err, "cannot add finalizer")
			return false, err
		}
		return true, nil
	}

	return false, nil
}
