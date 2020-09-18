package tupwas

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/tmax-cloud/l2c-operator/internal"
	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

const (
	IdePort    = 8080
	ReportPort = 80
	ConfigPort = 61436

	IdeVolume       = "git-report"
	IdeVolumeConfig = "config"

	ProjectDir = "/home/coder/project"
	ReportDir  = "/home/coder/.local/share/code-server/User/globalStorage/redhat.mta-vscode-extension/.mta/tooling/data/-38dkf89vj-wtx81drip"
)

func ideSecret(tupWas *tmaxv1.TupWAS) *corev1.Secret {
	password := utils.RandString(30)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ideReportResourceName(tupWas),
			Namespace: tupWas.Namespace,
			Labels:    ideReportLabels(tupWas),
		},
		StringData: map[string]string{
			"config.yaml": fmt.Sprintf(`bind-addr: 0.0.0.0:%d
auth: password
password: %s
cert: false`, IdePort, password),
			"password": password,
		},
	}
}

func ideReportService(tupWas *tmaxv1.TupWAS) (*corev1.Service, error) {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ideReportResourceName(tupWas),
			Namespace: tupWas.Namespace,
			Labels:    ideReportLabels(tupWas),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: "code",
					Port: IdePort,
				}, {
					Name: "report",
					Port: ReportPort,
				}, {
					Name: "config",
					Port: ConfigPort,
				},
			},
			Selector: ideReportServiceLabel(tupWas),
		},
	}, nil
}

func ideReportIngress(tupWas *tmaxv1.TupWAS) (*networkingv1beta1.Ingress, error) {
	return &networkingv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ideReportResourceName(tupWas),
			Namespace:   tupWas.Namespace,
			Labels:      ideReportLabels(tupWas),
			Annotations: tupWas.GenIngressAnnotation(),
		},
		Spec: networkingv1beta1.IngressSpec{
			Rules: []networkingv1beta1.IngressRule{{
				Host: IngressDefaultHost,
				IngressRuleValue: networkingv1beta1.IngressRuleValue{
					HTTP: &networkingv1beta1.HTTPIngressRuleValue{
						Paths: []networkingv1beta1.HTTPIngressPath{{
							Backend: networkingv1beta1.IngressBackend{
								ServiceName: ideReportResourceName(tupWas),
								ServicePort: intstr.IntOrString{Type: intstr.Int, IntVal: IdePort},
							},
						}},
					},
				},
			}, {
				Host: IngressDefaultHost,
				IngressRuleValue: networkingv1beta1.IngressRuleValue{
					HTTP: &networkingv1beta1.HTTPIngressRuleValue{
						Paths: []networkingv1beta1.HTTPIngressPath{{
							Backend: networkingv1beta1.IngressBackend{
								ServiceName: ideReportResourceName(tupWas),
								ServicePort: intstr.IntOrString{Type: intstr.Int, IntVal: ReportPort},
							},
						}},
					},
				},
			}, {
				Host: IngressDefaultHost,
				IngressRuleValue: networkingv1beta1.IngressRuleValue{
					HTTP: &networkingv1beta1.HTTPIngressRuleValue{
						Paths: []networkingv1beta1.HTTPIngressPath{{
							Backend: networkingv1beta1.IngressBackend{
								ServiceName: ideReportResourceName(tupWas),
								ServicePort: intstr.IntOrString{Type: intstr.Int, IntVal: ConfigPort},
							},
						}},
					},
				},
			}},
		},
	}, nil
}

func ideReportDeployment(tupWas *tmaxv1.TupWAS, configUrl, reportUrl string) (*appsv1.Deployment, error) {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ideReportResourceName(tupWas),
			Namespace: tupWas.Namespace,
			Labels:    ideReportLabels(tupWas),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ideReportServiceLabel(tupWas),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ideReportServiceLabel(tupWas),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            "ide",
						Image:           internal.EditorImage,
						ImagePullPolicy: corev1.PullAlways,
						Command:         []string{"/home/coder/scripts/init.sh"},
						Ports: []corev1.ContainerPort{{
							Name:          "code",
							ContainerPort: IdePort,
						}, {
							Name:          "config",
							ContainerPort: ConfigPort,
						}},
						Env: []corev1.EnvVar{{
							Name:  "PROJECT_ID",
							Value: tupWas.Name,
						}, {
							Name:  "CONFIG_URL",
							Value: configUrl,
						}, {
							Name:  "REPORT_URL",
							Value: reportUrl,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      IdeVolume,
							SubPath:   "project",
							MountPath: ProjectDir,
						}, {
							Name:      IdeVolume,
							SubPath:   "report",
							MountPath: ReportDir,
						}, {
							Name:      IdeVolumeConfig,
							SubPath:   "config.yaml",
							MountPath: "/home/coder/.config/code-server/config.yaml",
						}},
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: IdePort,
									},
								},
							},
						},
					}, {
						Name:            "report",
						Image:           "httpd:2.4",
						ImagePullPolicy: corev1.PullAlways,
						Ports: []corev1.ContainerPort{{
							Name:          "report",
							ContainerPort: ReportPort,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      IdeVolume,
							SubPath:   "report",
							MountPath: "/usr/local/apache2/htdocs/",
						}},
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: ReportPort,
									},
								},
							},
						},
					}},
					Volumes: []corev1.Volume{{
						Name: IdeVolume,
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: tupWas.GenResourceName(),
							},
						},
					}, {
						Name: IdeVolumeConfig,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{SecretName: ideReportResourceName(tupWas)},
						},
					}},
				},
			},
		},
	}, nil
}

func ideReportResourceName(tupWas *tmaxv1.TupWAS) string {
	return fmt.Sprintf("%s-ide", tupWas.Name)
}

func ideReportLabels(tupWas *tmaxv1.TupWAS) map[string]string {
	return map[string]string{
		"tupWas":    tupWas.Name,
		"component": "ide",
	}
}

func ideReportServiceLabel(tupWas *tmaxv1.TupWAS) map[string]string {
	return map[string]string{
		"tupWas": tupWas.Name,
		"tier":   "ide-pod",
	}
}
