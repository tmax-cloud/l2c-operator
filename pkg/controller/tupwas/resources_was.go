package tupwas

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func wasService(tupWas *tmaxv1.TupWAS) (*corev1.Service, error) {
	serviceType := tupWas.Spec.To.ServiceType
	// If service type is Ingress(=default), set it ClusterIP
	if serviceType == "" || serviceType == tmaxv1.WasServiceTypeIngress {
		serviceType = tmaxv1.WasServiceTypeClusterIP
	}
	port, err := tupWas.GenWasPort()
	if err != nil {
		return nil, err
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupWas.GenWasResourceName(),
			Namespace: tupWas.Namespace,
			Labels:    tupWas.GenWasLabels(),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceType(serviceType),
			Ports: []corev1.ServicePort{
				{
					Port: port,
				},
			},
			Selector: tupWas.GenWasServiceLabels(),
		},
	}, nil
}

func wasIngress(tupWas *tmaxv1.TupWAS) (*networkingv1beta1.Ingress, error) {
	port, err := tupWas.GenWasPort()
	if err != nil {
		return nil, err
	}
	return &networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        tupWas.GenWasResourceName(),
			Namespace:   tupWas.Namespace,
			Labels:      tupWas.GenWasLabels(),
			Annotations: tupWas.GenIngressAnnotation(),
		},
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{{
				Host: IngressDefaultHost,
				IngressRuleValue: networkingv1beta1.IngressRuleValue{
					HTTP: &networkingv1beta1.HTTPIngressRuleValue{
						Paths: []networkingv1beta1.HTTPIngressPath{{
							Backend: networkingv1beta1.IngressBackend{
								ServiceName: tupWas.GenWasResourceName(),
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

func wasDeploy(tupWas *tmaxv1.TupWAS) (*appsv1.Deployment, error) {
	port, err := tupWas.GenWasPort()
	if err != nil {
		return nil, err
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Labels: tupWas.GenWasLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: tupWas.GenWasServiceLabels(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: tupWas.GenWasServiceLabels(),
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

	if tupWas.Spec.To.Image.RegSecret != "" {
		dep.Spec.Template.Spec.ImagePullSecrets = append(dep.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: tupWas.Spec.To.Image.RegSecret})
	}

	return dep, nil
}
