package tupwas

import (
	"context"
	"fmt"
	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileTupWAS) deployWasNetwork(instance *tmaxv1.TupWAS) error {
	// Service for WAS deployment
	wasService, err := wasService(instance)
	if err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating service", err.Error()); err != nil {
			return err
		}
		return err
	}
	if err := utils.CheckAndCreateObject(wasService, nil, r.client, r.scheme, false); err != nil {
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating service", err.Error()); err != nil {
			return err
		}
		return err
	}

	// Ingress for WAS deployment - only if service type is Ingress(=default)
	if instance.Spec.To.ServiceType == "" || instance.Spec.To.ServiceType == tmaxv1.WasServiceTypeIngress {
		wasIngress, err := wasIngress(instance)
		if err != nil {
			if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating ingress", err.Error()); err != nil {
				return err
			}
			return err
		}
		if err := utils.CheckAndCreateObject(wasIngress, nil, r.client, r.scheme, false); err != nil {
			if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "error getting/creating ingress", err.Error()); err != nil {
				return err
			}
			return err
		}
	}

	return nil
}

func (r *ReconcileTupWAS) manageWasNetwork(instance *tmaxv1.TupWAS) error {
	// Update ingress - apply host
	wasIngress := &networkingv1beta1.Ingress{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.GenWasResourceName(), Namespace: instance.Namespace}, wasIngress); err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil {
		if len(wasIngress.Status.LoadBalancer.Ingress) != 0 && len(wasIngress.Spec.Rules) == 1 && wasIngress.Spec.Rules[0].Host == IngressDefaultHost {
			// If Loadbalancer is given to the ingress, but host is not set, set host!
			wasIngress.Spec.Rules[0].Host = fmt.Sprintf("%s.%s.%s.nip.io", instance.Name, instance.Namespace, wasIngress.Status.LoadBalancer.Ingress[0].IP)
			if err := r.client.Update(context.TODO(), wasIngress); err != nil {
				return err
			}
		}
	}

	// Update WAS URL
	wasUrl := ""
	wasService := &corev1.Service{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.GenWasResourceName(), Namespace: instance.Namespace}, wasService); err != nil && !errors.IsNotFound(err) {
		return err
	}
	switch instance.Spec.To.ServiceType {
	case tmaxv1.WasServiceTypeIngress, "":
		if len(wasIngress.Spec.Rules) == 1 && wasIngress.Spec.Rules[0].Host != IngressDefaultHost {
			wasUrl = fmt.Sprintf("http://%s", wasIngress.Spec.Rules[0].Host)
		}
	case tmaxv1.WasServiceTypeLoadBalancer:
		if len(wasService.Status.LoadBalancer.Ingress) > 0 && len(wasService.Spec.Ports) > 0 {
			wasUrl = fmt.Sprintf("http://%s:%d", wasService.Status.LoadBalancer.Ingress[0].IP, wasService.Spec.Ports[0].Port)
		}
	case tmaxv1.WasServiceTypeNodePort:
		nodes := &corev1.NodeList{}
		if err := r.client.List(context.TODO(), nodes); err != nil {
			return err
		}
		if len(nodes.Items) > 0 && len(wasService.Spec.Ports) > 0 {
			internalIp := ""
			for _, addr := range nodes.Items[0].Status.Addresses {
				if addr.Type == corev1.NodeInternalIP { // Should it be InternalIP..?
					internalIp = addr.Address
				}
			}
			if internalIp != "" {
				wasUrl = fmt.Sprintf("http://%s:%d", internalIp, wasService.Spec.Ports[0].NodePort)
			}
		}
	case tmaxv1.WasServiceTypeClusterIP:
		if len(wasService.Spec.Ports) > 0 {
			wasUrl = fmt.Sprintf("http://%s:%d", wasService.Spec.ClusterIP, wasService.Spec.Ports[0].Port)
		}
	default:
		if err := r.updateErrorStatus(instance, tmaxv1.WasConditionKeyProjectReady, corev1.ConditionFalse, "serviceType not supported", fmt.Sprintf("Service type %s is not supported", instance.Spec.To.ServiceType)); err != nil {
			return err
		}
		return nil
	}
	instance.Status.WasUrl = wasUrl

	return nil
}
