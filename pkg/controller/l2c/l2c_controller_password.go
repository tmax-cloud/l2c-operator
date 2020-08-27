package l2c

import (
	"context"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

func (r *ReconcileL2c) handlePassword(instance *tmaxv1.L2c) (bool, error) {
	changed := false

	if instance.Spec.Db == nil {
		return false, nil
	}

	// From PW
	if !utils.IsEncrypted(instance.Spec.Db.From.Password) {
		pw, err := utils.EncryptPassword(instance.Spec.Db.From.Password)
		if err != nil {
			return false, err
		}
		instance.Spec.Db.From.Password = pw
		changed = true
	}

	// To PW
	if !utils.IsEncrypted(instance.Spec.Db.To.Password) {
		pw, err := utils.EncryptPassword(instance.Spec.Db.To.Password)
		if err != nil {
			return false, err
		}
		instance.Spec.Db.To.Password = pw
		changed = true
	}

	// If changed, update spec
	if changed {
		if err := r.client.Update(context.TODO(), instance); err != nil {
			return false, err
		}
	}

	return changed, nil
}
