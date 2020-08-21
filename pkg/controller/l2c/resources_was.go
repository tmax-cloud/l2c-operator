package l2c

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func wasService(l2c *tmaxv1.L2c) (*corev1.Service, error) {
	port, err := wasPort(l2c)
	if err != nil {
		return nil, err
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wasResourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    wasLabels(l2c),
		},
		Spec: corev1.ServiceSpec{
			Type: "ClusterIP", // Should it be configurable? currently no...I think
			Ports: []corev1.ServicePort{
				{
					Port: port,
				},
			},
			Selector: wasServiceLabels(l2c),
		},
	}, nil
}

func wasIngress(l2c *tmaxv1.L2c) (*networkingv1beta1.Ingress, error) {
	port, err := wasPort(l2c)
	if err != nil {
		return nil, err
	}
	return &networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wasResourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    wasLabels(l2c),
		},
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{{
				Host: IngressDefaultHost,
				IngressRuleValue: networkingv1beta1.IngressRuleValue{
					HTTP: &networkingv1beta1.HTTPIngressRuleValue{
						Paths: []networkingv1beta1.HTTPIngressPath{{
							Backend: networkingv1beta1.IngressBackend{
								ServiceName: wasResourceName(l2c),
								ServicePort: intstr.IntOrString{
									Type:   intstr.Int,
									IntVal: port,
								},
							},
						}},
					},
				},
			}},
		},
	}, nil
}

func wasDeploy(l2c *tmaxv1.L2c) (*appsv1.Deployment, error) {
	port, err := wasPort(l2c)
	if err != nil {
		return nil, err
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels: wasLabels(l2c),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: wasServiceLabels(l2c),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: wasServiceLabels(l2c),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Ports: []corev1.ContainerPort{{
							ContainerPort: port,
						}},
					}},
				},
			},
		},
	}

	if l2c.Spec.Was.To.Image.RegSecret != "" {
		dep.Spec.Template.Spec.ImagePullSecrets = append(dep.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: l2c.Spec.Was.To.Image.RegSecret})
	}

	return dep, nil
}

// Supporting functions
func wasResourceName(l2c *tmaxv1.L2c) string {
	return fmt.Sprintf("%s-was", l2c.Name)
}

func wasLabels(l2c *tmaxv1.L2c) map[string]string {
	return map[string]string{
		"l2c":       l2c.Name,
		"component": "was",
	}
}

func wasServiceLabels(l2c *tmaxv1.L2c) map[string]string {
	return map[string]string{
		"l2c":  l2c.Name,
		"tier": l2c.Spec.Was.To.Type,
	}
}

func wasPort(l2c *tmaxv1.L2c) (int32, error) {
	switch l2c.Spec.Was.To.Type {
	case "jeus":
		return 8808, nil
	default:
		return 0, fmt.Errorf("spec.was.to.type(%s) not supported", l2c.Spec.Was.To.Type)
	}
}
