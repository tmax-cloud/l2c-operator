package sonarqube

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (s *SonarQube) deployment() *appsv1.Deployment {
	volumeName := "sonar-pv"
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ResourceName,
			Namespace: s.Namespace,
			Labels:    label,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: depLabel,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: depLabel,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           s.Image,
							ImagePullPolicy: corev1.PullAlways,
							Name:            "sonarqube",
							Ports: []corev1.ContainerPort{
								{
									Name:          "web",
									ContainerPort: 9000,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      volumeName,
									MountPath: "/opt/sonarqube/data",
									SubPath:   "data",
								},
								{
									Name:      volumeName,
									MountPath: "/opt/sonarqube/logs",
									SubPath:   "logs",
								},
							},
							ReadinessProbe: &corev1.Probe{
								PeriodSeconds:    1,
								SuccessThreshold: 5,
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/api/users/search",
										Port: intstr.IntOrString{IntVal: 9000},
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: volumeName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: s.ResourceName,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (s *SonarQube) service() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ResourceName,
			Namespace: s.Namespace,
			Labels:    label,
		},
		Spec: corev1.ServiceSpec{
			Type:     "ClusterIP",
			Selector: depLabel,
			Ports: []corev1.ServicePort{
				{
					Port: 9000,
				},
			},
		},
	}
}

func (s *SonarQube) pvc() *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ResourceName,
			Namespace: s.Namespace,
			Labels:    label,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &s.StorageClassName,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: resource.MustParse(s.StorageSize),
				},
			},
		},
	}
}

func (s *SonarQube) secret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ResourceName,
			Namespace: s.Namespace,
			Labels:    label,
		},
		StringData: map[string]string{
			SecretKeyAdminId:       DefaultAdminId,
			SecretKeyAdminPw:       DefaultAdminPw,
			SecretKeyToken:         "",
			SecretKeyAnalyzerId:    DefaultAnalyzerId,
			SecretKeyAnalyzerPw:    DefaultAnalyzerPw,
			SecretKeyAnalyzerToken: "",
		},
	}
}
