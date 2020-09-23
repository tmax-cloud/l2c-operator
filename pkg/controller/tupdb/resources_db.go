package tupdb

import (
	"fmt"
	"github.com/tmax-cloud/l2c-operator/internal/utils"
	"strconv"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tmax-cloud/l2c-operator/internal"
)

const (
	DbVolumeName   = "db-volume"
	IngressDefault = "waiting.for.ingress.ready"
)

func dbPvc(dbInstance *tmaxv1.TupDB) (*corev1.PersistentVolumeClaim, error) {
	storageQuantity, err := resource.ParseQuantity(dbInstance.Spec.To.StorageSize)
	if err != nil {
		return nil, err
	}
	return &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbResourceName(dbInstance),
			Namespace: dbInstance.Namespace,
			Labels:    dbLabels(dbInstance),
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

func dbService(dbInstance *tmaxv1.TupDB) (*corev1.Service, error) {
	port, err := dbPort(dbInstance)
	if err != nil {
		return nil, err
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbResourceName(dbInstance),
			Namespace: dbInstance.Namespace,
			Labels:    dbLabels(dbInstance),
		},
		Spec: corev1.ServiceSpec{
			Type: "LoadBalancer", // Should it be configurable? currently no...I think
			Ports: []corev1.ServicePort{
				{
					Port: port,
				},
			},
			Selector: dbServiceLabels(dbInstance),
		},
	}, nil
}

func tupDBSecret(dbInstance *tmaxv1.TupDB) (*corev1.Secret, error) {
	logger := utils.NewTupLogger(tmaxv1.TupDB{}, dbInstance.Namespace, dbInstance.Name)
	secretVal, err := tupDbSecretValues(dbInstance)

	if err != nil {
		logger.Error(err, "Db Secret Error")
		return nil, err
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      tmaxv1.TupDBSecretName,
			Namespace: dbInstance.Namespace,
		},
		StringData: secretVal,
	}, nil
}

func dbDeploySecret(dbInstance *tmaxv1.TupDB) (*corev1.Secret, error) {
	logger := utils.NewTupLogger(tmaxv1.TupDB{}, dbInstance.Namespace, dbInstance.Name)
	secretVal, err := dbSecretValues(dbInstance)
	if err != nil {
		logger.Error(err, "Db Secret Error")
		return nil, err
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbResourceName(dbInstance),
			Namespace: dbInstance.Namespace,
			Labels:    dbLabels(dbInstance),
		},
		StringData: secretVal,
	}, nil
}

func dbDeploy(dbInstance *tmaxv1.TupDB) (*appsv1.Deployment, error) {
	cont, err := dbContainer(dbInstance)
	if err != nil {
		return nil, err
	}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbResourceName(dbInstance),
			Namespace: dbInstance.Namespace,
			Labels:    dbLabels(dbInstance),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: dbServiceLabels(dbInstance),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: dbServiceLabels(dbInstance),
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
									ClaimName: dbResourceName(dbInstance),
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
func dbResourceName(dbInstance *tmaxv1.TupDB) string {
	return fmt.Sprintf("%s-db", dbInstance.Name)
}

func dbLabels(dbInstance *tmaxv1.TupDB) map[string]string {
	return map[string]string{
		"tupDB": dbInstance.Name,
	}
}

func dbServiceLabels(dbInstance *tmaxv1.TupDB) map[string]string {
	return map[string]string{
		"tupDB": dbInstance.Name,
	}
}

func dbPort(dbInstance *tmaxv1.TupDB) (int32, error) {
	switch dbInstance.Spec.To.Type {
	case tmaxv1.DbTypeTibero:
		return 8629, nil
	default:
		return 0, fmt.Errorf("spec.db.to.type(%s) not supported", dbInstance.Spec.To.Type)
	}
}

func tupDbSecretValues(dbInstance *tmaxv1.TupDB) (map[string]string, error) {
	// [TODO] Decrypt Password
	//pw, err := utils.DecryptPassword(dbInstance.Spec.To.Password)
	//if err != nil {
	//	return nil, err
	//}

	values := map[string]string{}
	values["source-user"] = dbInstance.Spec.From.User
	values["source-password"] = dbInstance.Spec.From.Password
	values["source-sid"] = dbInstance.Spec.From.Sid
	values["target-user"] = dbInstance.Spec.To.User
	values["target-password"] = dbInstance.Spec.To.Password
	values["target-sid"] = dbInstance.Spec.To.Sid

	return values, nil
}

func dbSecretValues(dbInstance *tmaxv1.TupDB) (map[string]string, error) {
	port, err := dbPort(dbInstance)
	if err != nil {
		return nil, err
	}

	// [TODO] Decrypt Password
	//pw, err := utils.DecryptPassword(dbInstance.Spec.To.Password)
	//if err != nil {
	//	return nil, err
	//}

	values := map[string]string{}
	switch dbInstance.Spec.To.Type {
	case tmaxv1.DbTypeTibero:
		values["MASTER_USER"] = dbInstance.Spec.To.User
		values["MASTER_PASSWORD"] = dbInstance.Spec.To.Password
		values["TCS_INSTALL"] = "1"
		values["TCS_SID"] = dbInstance.Spec.To.User
		values["TB_SID"] = dbInstance.Spec.To.User
		values["TCS_PORT"] = strconv.Itoa(int(port))
	default:
		return nil, fmt.Errorf("spec.db.to.type(%s) not supported", dbInstance.Spec.To.Type)
	}

	return values, nil
}

func dbContainer(dbInstance *tmaxv1.TupDB) (*corev1.Container, error) {
	port, err := dbPort(dbInstance)
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

	switch dbInstance.Spec.To.Type {
	case tmaxv1.DbTypeTibero:
		cont.Image = "192.168.6.110:5000/cloud_tcs_tibero_standalone:200309"

		// Set env.s
		envKeys := []string{"MASTER_USER", "MASTER_PASSWORD", "TCS_INSTALL", "TCS_SID", "TB_SID", "TCS_PORT"}
		for _, k := range envKeys {
			cont.Env = append(cont.Env, corev1.EnvVar{
				Name: k,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: dbResourceName(dbInstance)},
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
		return nil, fmt.Errorf("spec.db.to.type(%s) not supported", dbInstance.Spec.To.Type)
	}

	return cont, nil
}
