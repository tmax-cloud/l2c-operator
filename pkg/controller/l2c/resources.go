package l2c

import (
	"bytes"
	"fmt"
	"strings"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/tmax-cloud/l2c-operator/internal"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

// ConfigMap for DB deployment
// This does not belong to resources_db 'cause it's for L2c process itself, not for DB runtime
func dbConfigMap(l2c *tmaxv1.L2c) (*corev1.ConfigMap, error) {
	if l2c.Spec.Db == nil {
		return nil, fmt.Errorf("db migration is not configured")
	}

	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
		Yaml:   true,
		Pretty: true,
		Strict: true,
	})

	// PVC object
	pvc, err := dbPvc(l2c)
	if err != nil {
		return nil, err
	}
	pvcBuf := new(bytes.Buffer)
	if err := serializer.Encode(pvc, pvcBuf); err != nil {
		return nil, err
	}

	// Service object
	svc, err := dbService(l2c)
	if err != nil {
		return nil, err
	}
	svcBuf := new(bytes.Buffer)
	if err := serializer.Encode(svc, svcBuf); err != nil {
		return nil, err
	}

	// Secret object
	secret, err := dbSecret(l2c)
	if err != nil {
		return nil, err
	}
	secretBuf := new(bytes.Buffer)
	if err := serializer.Encode(secret, secretBuf); err != nil {
		return nil, err
	}

	// Deployment object
	deploy, err := dbDeploy(l2c)
	if err != nil {
		return nil, err
	}
	deployBuf := new(bytes.Buffer)
	if err := serializer.Encode(deploy, deployBuf); err != nil {
		return nil, err
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dbResourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    labels(l2c),
		},
		Data: map[string]string{
			tmaxv1.DbConfigMapKeyPvc:    pvcBuf.String(),
			tmaxv1.DbConfigMapKeySvc:    svcBuf.String(),
			tmaxv1.DbConfigMapKeySecret: secretBuf.String(),
			tmaxv1.DbConfigMapKeyDeploy: deployBuf.String(),
		},
	}, nil
}

// ConfigMap for WAS deployment spec
// This does not belong to resources_Was 'cause it's for L2c process itself, not for WAS runtime
func wasConfigMap(l2c *tmaxv1.L2c) (*corev1.ConfigMap, error) {
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
		Yaml:   true,
		Pretty: true,
		Strict: true,
	})

	// Deployment object
	deploy, err := wasDeploy(l2c)
	if err != nil {
		return nil, err
	}
	deployBuf := new(bytes.Buffer)
	if err := serializer.Encode(deploy, deployBuf); err != nil {
		return nil, err
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wasResourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    labels(l2c),
		},
		Data: map[string]string{
			"deploy-spec.yaml": deployBuf.String(),
		},
	}, nil
}

func secret(l2c *tmaxv1.L2c) (*corev1.Secret, error) {
	if l2c.Spec.Db == nil {
		return nil, fmt.Errorf("db migration is not configured")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    labels(l2c),
		},
		StringData: map[string]string{
			tmaxv1.DbSecretKeySourceUser:     l2c.Spec.Db.From.User,
			tmaxv1.DbSecretKeySourcePassword: l2c.Spec.Db.From.Password,
			tmaxv1.DbSecretKeySourceSid:      l2c.Spec.Db.From.Sid,
			tmaxv1.DbSecretKeyTargetUser:     l2c.Spec.Db.To.User,
			tmaxv1.DbSecretKeyTargetPassword: l2c.Spec.Db.To.Password,
			tmaxv1.DbSecretKeyTargetSid:      l2c.Spec.Db.To.User,
		},
	}, nil
}

func serviceAccount(l2c *tmaxv1.L2c) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    labels(l2c),
		},
	}
}

func roleBinding(l2c *tmaxv1.L2c) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    labels(l2c),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "l2c",
		},
		Subjects: []rbacv1.Subject{{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      l2c.Name,
			Namespace: l2c.Namespace,
		}},
	}
}

func pipeline(l2c *tmaxv1.L2c) (*tektonv1.Pipeline, error) {
	// doMigrateDb
	doMigrateDb := "TRUE"
	if l2c.Spec.Db == nil {
		doMigrateDb = "FALSE"
	}

	// DB port
	var dbPortNum int32 = 0
	if l2c.Spec.Db != nil {
		var err error
		dbPortNum, err = dbPort(l2c)
		if err != nil {
			return nil, err
		}
	}

	// Builder Image
	builderImg, err := builderImage(l2c)
	if err != nil {
		return nil, err
	}

	return &tektonv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName(l2c),
			Namespace: l2c.Namespace,
			Labels:    labels(l2c),
		},
		Spec: tektonv1.PipelineSpec{
			Resources: []tektonv1.PipelineDeclaredResource{{
				Name: string(tmaxv1.PipelineResourceNameGit),
				Type: tektonv1.PipelineResourceTypeGit,
			}, {
				Name: string(tmaxv1.PipelineResourceNameImage),
				Type: tektonv1.PipelineResourceTypeImage,
			}},
			Params: []tektonv1.ParamSpec{
				{Name: tmaxv1.PipelineParamNameSonarUrl},
				{Name: tmaxv1.PipelineParamNameSonarToken},
				{Name: tmaxv1.PipelineParamNameSonarProjectKey},
			},
			Tasks: []tektonv1.PipelineTask{{
				Name:    string(tmaxv1.PipelineTaskNameAnalyze),
				TaskRef: &tektonv1.TaskRef{Name: tmaxv1.TaskNameAnalyzeMaven, Kind: tektonv1.ClusterTaskKind}, // TODO: MAVEN/GRADLE
				Resources: &tektonv1.PipelineTaskResources{
					Inputs: []tektonv1.PipelineTaskInputResource{{
						Name:     string(tmaxv1.PipelineResourceNameGit),
						Resource: string(tmaxv1.PipelineResourceNameGit),
					}},
				},
				Params: []tektonv1.Param{{
					Name:  tmaxv1.PipelineParamNameSonarUrl,
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.PipelineParamNameSonarUrl)},
				}, {
					Name:  tmaxv1.PipelineParamNameSonarToken,
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.PipelineParamNameSonarToken)},
				}, {
					Name:  tmaxv1.PipelineParamNameSonarProjectKey,
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.PipelineParamNameSonarProjectKey)},
				}},
			}, {
				Name:    string(tmaxv1.PipelineTaskNameMigrate),
				TaskRef: &tektonv1.TaskRef{Name: tmaxv1.TaskNameDbMigration, Kind: tektonv1.ClusterTaskKind},
				Params: []tektonv1.Param{{
					Name:  "DO_MIGRATE_DB",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: doMigrateDb},
				}, {
					Name:  "CM_NAME",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: dbResourceName(l2c)},
				}, {
					Name:  "SECRET_NAME",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: resourceName(l2c)},
				}, {
					Name:  "SOURCE_TYPE",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: strings.ToUpper(l2c.Spec.Db.From.Type)},
				}, {
					Name:  "SOURCE_HOST",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: l2c.Spec.Db.From.Host},
				}, {
					Name:  "SOURCE_PORT",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("%d", l2c.Spec.Db.From.Port)},
				}, {
					Name:  "TARGET_TYPE",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: strings.ToUpper(l2c.Spec.Db.To.Type)},
				}, {
					Name:  "TARGET_HOST",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: dbResourceName(l2c)}, // Host : service for DB deployment
				}, {
					Name:  "TARGET_PORT",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("%d", dbPortNum)},
				}},
				RunAfter: []string{string(tmaxv1.PipelineTaskNameAnalyze)},
			}, {
				Name:    string(tmaxv1.PipelineTaskNameBuild),
				TaskRef: &tektonv1.TaskRef{Name: tmaxv1.TaskNameBuild, Kind: tektonv1.ClusterTaskKind},
				Resources: &tektonv1.PipelineTaskResources{
					Inputs: []tektonv1.PipelineTaskInputResource{{
						Name:     "source",
						Resource: string(tmaxv1.PipelineResourceNameGit),
					}},
					Outputs: []tektonv1.PipelineTaskOutputResource{{
						Name:     "image",
						Resource: string(tmaxv1.PipelineResourceNameImage),
					}},
				},
				Params: []tektonv1.Param{{
					Name:  "BUILDER_IMAGE",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: builderImg},
				}, {
					Name:  "REGISTRY_SECRET_NAME",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: l2c.Spec.Was.To.Image.RegSecret},
				}},
				RunAfter: []string{string(tmaxv1.PipelineTaskNameMigrate)},
			}, {
				Name:    string(tmaxv1.PipelineTaskNameDeploy),
				TaskRef: &tektonv1.TaskRef{Name: tmaxv1.TaskNameDeploy, Kind: tektonv1.ClusterTaskKind},
				Resources: &tektonv1.PipelineTaskResources{
					Inputs: []tektonv1.PipelineTaskInputResource{{
						Name:     "image",
						Resource: string(tmaxv1.PipelineResourceNameImage),
					}},
				},
				Params: []tektonv1.Param{{
					Name:  "app-name",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: l2c.Name},
				}, {
					Name:  "image-url",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(tasks.%s.results.image-url)", string(tmaxv1.PipelineTaskNameBuild))},
				}, {
					Name:  "deploy-cfg-name",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: wasResourceName(l2c)},
				}},
				RunAfter: []string{string(tmaxv1.PipelineTaskNameBuild)},
			}},
		},
	}, nil
}

// Supporting functions
func resourceName(l2c *tmaxv1.L2c) string {
	return l2c.Name
}

func labels(l2c *tmaxv1.L2c) map[string]string {
	return map[string]string{
		"l2c":       l2c.Name,
		"component": "l2c",
	}
}

func builderImage(l2c *tmaxv1.L2c) (string, error) {
	switch l2c.Spec.Was.To.Type {
	case "jeus":
		return internal.BuilderImageJeus, nil
	default:
		return "", fmt.Errorf("%s was type is not supported", l2c.Spec.Was.To.Type)
	}
}
