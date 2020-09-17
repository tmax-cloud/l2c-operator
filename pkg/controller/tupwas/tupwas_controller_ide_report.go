package tupwas

import (
	"context"
	"fmt"
	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	IdePrefix    = "ide"
	ReportPrefix = "report"
	ConfigPrefix = "config"
)

func (r *ReconcileTupWAS) deployIdeReport(instance *tmaxv1.TupWAS) error {
	// Generate VSCode - Secret/Service/Ingress/Deployment
	idePassword := []byte("")
	ideUrl := ""
	reportUrl := ""

	// Generate Secret
	ideSecret := ideSecret(instance)
	if err := r.createAndUpdateStatus(ideSecret, instance, "error getting/creating pipeline"); err != nil {
		return err
	}
	// Check IDE Password
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: ideSecret.Name, Namespace: ideSecret.Namespace}, ideSecret)
	if err != nil && !errors.IsNotFound(err) {
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating secret", err.Error()); err != nil {
			return err
		}
		return err
	}
	if err == nil {
		idePassword = ideSecret.Data["password"]
	}

	// Generate Service
	ideService, err := ideReportService(instance)
	if err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating service", err.Error()); err != nil {
			return err
		}
		return err
	}
	if err := r.createAndUpdateStatus(ideService, instance, "error getting/creating pipeline"); err != nil {
		return err
	}

	// Generate Ingress
	ideIngress, err := ideReportIngress(instance)
	if err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating ingress", err.Error()); err != nil {
			return err
		}
		return err
	}
	if err := r.createAndUpdateStatus(ideIngress, instance, "error getting/creating pipeline"); err != nil {
		return err
	}

	// Check ingress status first before deploy
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ideIngress.Name, Namespace: ideIngress.Namespace}, ideIngress)
	if err != nil && !errors.IsNotFound(err) {
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating ingress", err.Error()); err != nil {
			return err
		}
		return err
	} else if err != nil && errors.IsNotFound(err) {
		if err := r.setCondition(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "ingress for ide is not ready", err.Error()); err != nil {
			return err
		}
	} else if err == nil {
		if len(ideIngress.Status.LoadBalancer.Ingress) == 0 {
			if err := r.setCondition(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "ingress for ide is not ready", "ingress didn't get external ip yet"); err != nil {
				return err
			}
		} else if len(ideIngress.Status.LoadBalancer.Ingress) != 0 && len(ideIngress.Spec.Rules) == 3 && (ideIngress.Spec.Rules[0].Host == IngressDefaultHost || ideIngress.Spec.Rules[1].Host == IngressDefaultHost || ideIngress.Spec.Rules[2].Host == IngressDefaultHost) {
			// If Loadbalancer is given to the ingress, but host is not set, set host!
			ideIngress.Spec.Rules[0].Host = fmt.Sprintf("%s.%s.%s.%s.nip.io", IdePrefix, instance.Name, instance.Namespace, ideIngress.Status.LoadBalancer.Ingress[0].IP)
			ideIngress.Spec.Rules[1].Host = fmt.Sprintf("%s.%s.%s.%s.nip.io", ReportPrefix, instance.Name, instance.Namespace, ideIngress.Status.LoadBalancer.Ingress[0].IP)
			ideIngress.Spec.Rules[2].Host = fmt.Sprintf("%s.%s.%s.%s.nip.io", ConfigPrefix, instance.Name, instance.Namespace, ideIngress.Status.LoadBalancer.Ingress[0].IP)
			if err := r.client.Update(context.TODO(), ideIngress); err != nil {
				return err
			}
		} else if len(ideIngress.Spec.Rules) == 3 && ideIngress.Spec.Rules[0].Host != IngressDefaultHost && ideIngress.Spec.Rules[1].Host != IngressDefaultHost && ideIngress.Spec.Rules[2].Host != IngressDefaultHost {
			// Update ingress url to a status field
			ideUrl = fmt.Sprintf("http://%s", ideIngress.Spec.Rules[0].Host)
			reportUrl = fmt.Sprintf("http://%s", ideIngress.Spec.Rules[1].Host)
		}

		// Generate Deployment only if ingress is ready
		if len(ideIngress.Spec.Rules) == 3 && ideIngress.Spec.Rules[0].Host != IngressDefaultHost && ideIngress.Spec.Rules[1].Host != IngressDefaultHost && ideIngress.Spec.Rules[2].Host != IngressDefaultHost {
			ideDeploy, err := ideReportDeployment(instance, ideIngress.Spec.Rules[2].Host, ideIngress.Spec.Rules[1].Host)
			if err != nil {
				return err
			}
			if err := utils.CheckAndCreateObject(ideDeploy, instance, r.client, r.scheme, false); err != nil {
				return err
			}

			// Not ready if Deployment has non-ready container
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: ideDeploy.Name, Namespace: ideDeploy.Namespace}, ideDeploy)
			if err != nil && !errors.IsNotFound(err) {
				if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating deployment", err.Error()); err != nil {
					return err
				}
				return err
			} else if (err != nil && errors.IsNotFound(err)) || (err == nil && (ideDeploy.Status.Replicas == 0 || ideDeploy.Status.Replicas != ideDeploy.Status.ReadyReplicas)) {
				msg := "some replicas are not ready yet"
				if err != nil {
					msg = err.Error()
				}
				if err := r.setCondition(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "deployment for ide is not ready", msg); err != nil {
					return err
				}
			}
		}
	}

	// Save it to status - only if analyze is complete
	if instance.Status.LastAnalyzeCompletionTime != nil {
		if instance.Status.Editor == nil {
			instance.Status.Editor = &tmaxv1.EditorStatus{}
		}
		instance.Status.Editor.Password = string(idePassword)
		instance.Status.Editor.Url = ideUrl
		instance.Status.ReportUrl = reportUrl
	}

	return nil
}
