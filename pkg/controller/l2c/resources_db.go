package l2c

import (
	"fmt"
	"strconv"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tmax-cloud/l2c-operator/internal"
)

const (
	DbVolumeName = "db-volume"
)

// TODO: These resources should be configured using configmap or something else! not in this code!
func dbPvc(l2c *tmaxv1.L2c) (*corev1.PersistentVolumeClaim, error) {
	storageQuantity, err := resource.ParseQuantity(l2c.Spec.Db.To.StorageSize)
	if err != nil {
		return nil, err
	}
	return &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbResourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    dbLabels(l2c),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &internal.StorageClassName,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": storageQuantity,
				},
			},
		},
	}, nil
}

func dbService(l2c *tmaxv1.L2c) (*corev1.Service, error) {
	port, err := dbPort(l2c)
	if err != nil {
		return nil, err
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbResourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    dbLabels(l2c),
		},
		Spec: corev1.ServiceSpec{
			Type: "ClusterIP", // Should it be configurable? currently no...I think
			Ports: []corev1.ServicePort{
				{
					Port: port,
				},
			},
			Selector: dbServiceLabels(l2c),
		},
	}, nil
}

func dbSecret(l2c *tmaxv1.L2c) (*corev1.Secret, error) {
	secretVal, err := dbSecretValues(l2c)
	if err != nil {
		return nil, err
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbResourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    dbLabels(l2c),
		},
		StringData: secretVal,
	}, nil
}

func dbDeploy(l2c *tmaxv1.L2c) (*appsv1.Deployment, error) {
	cont, err := dbContainer(l2c)
	if err != nil {
		return nil, err
	}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbResourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    dbLabels(l2c),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: dbServiceLabels(l2c),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: dbServiceLabels(l2c),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						*cont,
					},
					Volumes: []corev1.Volume{
						{
							Name: DbVolumeName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: dbResourceName(l2c),
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

// Supporting functions
func dbResourceName(l2c *tmaxv1.L2c) string {
	return fmt.Sprintf("%s-db", l2c.Name)
}

func dbLabels(l2c *tmaxv1.L2c) map[string]string {
	return map[string]string{
		"l2c":       l2c.Name,
		"component": "db",
	}
}

func dbServiceLabels(l2c *tmaxv1.L2c) map[string]string {
	return map[string]string{
		"l2c":  l2c.Name,
		"tier": l2c.Spec.Db.To.Type,
	}
}

func dbPort(l2c *tmaxv1.L2c) (int32, error) {
	switch l2c.Spec.Db.To.Type {
	case "tibero":
		return 8629, nil
	default:
		return 0, fmt.Errorf("spec.db.to.type(%s) not supported", l2c.Spec.Db.To.Type)
	}
}

func dbSecretValues(l2c *tmaxv1.L2c) (map[string]string, error) {
	port, err := dbPort(l2c)
	if err != nil {
		return nil, err
	}

	values := map[string]string{}
	switch l2c.Spec.Db.To.Type {
	case "tibero":
		values["MASTER_USER"] = l2c.Spec.Db.To.User
		values["MASTER_PASSWORD"] = l2c.Spec.Db.To.Password
		values["TCS_INSTALL"] = "1"
		values["TCS_SID"] = l2c.Spec.Db.To.User
		values["TB_SID"] = l2c.Spec.Db.To.User
		values["TCS_PORT"] = strconv.Itoa(int(port))
	default:
		return nil, fmt.Errorf("spec.db.to.type(%s) not supported", l2c.Spec.Db.To.Type)
	}

	return values, nil
}

func dbContainer(l2c *tmaxv1.L2c) (*corev1.Container, error) {
	port, err := dbPort(l2c)
	if err != nil {
		return nil, err
	}
	cont := &corev1.Container{
		Name: "database",
		Ports: []corev1.ContainerPort{
			{
				Name:          "database",
				ContainerPort: port,
			},
		},
	}

	switch l2c.Spec.Db.To.Type {
	case "tibero":
		cont.Image = "192.168.6.110:5000/cloud_tcs_tibero_standalone:200309"

		// Set env.s
		envKeys := []string{"MASTER_USER", "MASTER_PASSWORD", "TCS_INSTALL", "TCS_SID", "TB_SID", "TCS_PORT"}
		for _, k := range envKeys {
			cont.Env = append(cont.Env, corev1.EnvVar{
				Name: k,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: dbResourceName(l2c)},
						Key:                  k,
					},
				},
			})
		}

		cont.VolumeMounts = append(cont.VolumeMounts, corev1.VolumeMount{
			Name:      DbVolumeName,
			MountPath: "/tibero/mnt/tibero",
		})

		cont.Lifecycle = &corev1.Lifecycle{
			PostStart: &corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"/bin/bash",
						"-c",
						`echo 'SELECT COUNT(*) FROM all_tables;' > /tmp/test.sql
echo 'EXIT;' >> /tmp/test.sql
echo "#!/bin/bash" > /tmp/probe.sh
echo "TEST=\$(tbsql $MASTER_USER/$MASTER_PASSWORD @/tmp/test.sql | grep -E '[0-9]* row[s]? selected')" >> /tmp/probe.sh
echo "[ \"\$TEST\" == \"\" ] && exit 1 || exit 0" >> /tmp/probe.sh
chmod +x /tmp/probe.sh`,
					},
				},
			},
		}

		cont.ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 5,
			PeriodSeconds:       10,
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"/bin/bash",
						"-c",
						"/tmp/probe.sh",
					},
				},
			},
		}
	default:
		return nil, fmt.Errorf("spec.db.to.type(%s) not supported", l2c.Spec.Db.To.Type)
	}

	return cont, nil
}
