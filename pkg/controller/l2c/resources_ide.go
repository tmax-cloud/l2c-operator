package l2c

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	"github.com/tmax-cloud/l2c-operator/pkg/sonarqube"
)

const (
	IdePort         = 8080
	IdeVolumeConfig = "config"
)

func ideSecret(l2c *tmaxv1.L2c, sonar *sonarqube.SonarQube) *corev1.Secret {
	password := utils.RandString(30)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ideResourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    ideLabels(l2c),
		},
		StringData: map[string]string{
			"settings.json": fmt.Sprintf(`{
        "sonarlint.connectedMode.connections.sonarqube": [
          {
            "serverUrl": "%s/",
            "token": "%s"
          }
        ],
        "sonarlint.connectedMode.project": {
          "projectKey": "%s"
        },
        "java.semanticHighlighting.enabled": true,
        "sonarlint.ls.javaHome": "/usr/lib/jvm/java-11-openjdk-amd64",
        "java.home": "/usr/lib/jvm/java-11-openjdk-amd64"
      }
`, sonar.URL, sonar.AnalyzerToken, l2c.GetSonarProjectName()),
			"config.yaml": fmt.Sprintf(`bind-addr: 0.0.0.0:%d
auth: password
password: %s
cert: false`, IdePort, password),
			"password": password,
		},
	}
}

func ideService(l2c *tmaxv1.L2c) (*corev1.Service, error) {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ideResourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    ideLabels(l2c),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: IdePort,
				},
			},
			Selector: ideServiceLabel(l2c),
		},
	}, nil
}

func ideIngress(l2c *tmaxv1.L2c) (*networkingv1beta1.Ingress, error) {
	return &networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ideResourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    ideLabels(l2c),
		},
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{{
				Host: IngressDefaultHost,
				IngressRuleValue: networkingv1beta1.IngressRuleValue{
					HTTP: &networkingv1beta1.HTTPIngressRuleValue{
						Paths: []networkingv1beta1.HTTPIngressPath{{
							Backend: networkingv1beta1.IngressBackend{
								ServiceName: ideResourceName(l2c),
								ServicePort: intstr.IntOrString{
									Type:   intstr.Int,
									IntVal: IdePort,
								},
							},
						}},
					},
				},
			}},
		},
	}, nil
}

func ideDeployment(l2c *tmaxv1.L2c) (*appsv1.Deployment, error) {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ideResourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    ideLabels(l2c),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ideServiceLabel(l2c),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ideServiceLabel(l2c),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            "web-ide",
						Image:           "192.168.6.110:5000/tmax/code-server:3.3.1", // TODO!!!!
						Args:            []string{""},
						ImagePullPolicy: corev1.PullAlways,
						VolumeMounts: []corev1.VolumeMount{{
							Name:      IdeVolumeConfig,
							SubPath:   "settings.json",
							MountPath: "/tmp/settings.json",
						}, {
							Name:      IdeVolumeConfig,
							SubPath:   "config.yaml",
							MountPath: "/home/coder/.config/code-server/config.yaml",
						}},
						Lifecycle: &corev1.Lifecycle{
							PostStart: &corev1.Handler{
								Exec: &corev1.ExecAction{
									Command: []string{
										"/bin/bash",
										"-c",
										fmt.Sprintf("git clone %s ~/project/%s; cp /tmp/settings.json /home/coder/.local/share/code-server/User/settings.json", l2c.Spec.Was.From.Git.Url, l2c.Name),
									},
								},
							},
						},
					}},
					Volumes: []corev1.Volume{{
						Name: IdeVolumeConfig,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{SecretName: ideResourceName(l2c)},
						},
					}},
				},
			},
		},
	}, nil
}

func ideResourceName(l2c *tmaxv1.L2c) string {
	return fmt.Sprintf("%s-ide", l2c.Name)
}

func ideLabels(l2c *tmaxv1.L2c) map[string]string {
	return map[string]string{
		"l2c":       l2c.Name,
		"component": "ide",
	}
}

func ideServiceLabel(l2c *tmaxv1.L2c) map[string]string {
	return map[string]string{
		"l2c":  l2c.Name,
		"tier": "ide-pod",
	}
}
