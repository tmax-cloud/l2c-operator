package l2c

import (
	"bytes"
	"fmt"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

// ConfigMap for DB deployment
func configMap(l2c *tmaxv1.L2c) (*corev1.ConfigMap, error) {
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
	svc, err := dbSvc(l2c)
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
			Name:      l2c.Name,
			Namespace: l2c.Namespace,
		},
		Data: map[string]string{
			tmaxv1.ConfigMapKeyPvc:    pvcBuf.String(),
			tmaxv1.ConfigMapKeySvc:    svcBuf.String(),
			tmaxv1.ConfigMapKeySecret: secretBuf.String(),
			tmaxv1.ConfigMapKeyDeploy: deployBuf.String(),
		},
	}, nil
}

func pipeline(l2c *tmaxv1.L2c) *tektonv1.Pipeline {
	return &tektonv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      l2c.Name,
			Namespace: l2c.Namespace,
		},
		Spec: tektonv1.PipelineSpec{
			Description: "dummy pipeline",
			Resources: []tektonv1.PipelineDeclaredResource{
				{
					Name: tmaxv1.PipelineResourceNameGit,
					Type: tektonv1.PipelineResourceTypeGit,
				},
			},
			Params: []tektonv1.ParamSpec{
				{Name: tmaxv1.PipelineParamNameSonarUrl},
				{Name: tmaxv1.PipelineParamNameSonarToken},
				{Name: tmaxv1.PipelineParamNameSonarProjectKey},
			},
			Tasks: []tektonv1.PipelineTask{
				{
					Name:    string(tmaxv1.PipelineTaskNameAnalyze),
					TaskRef: &tektonv1.TaskRef{Name: tmaxv1.TaskNameAnalyzeMaven, Kind: tektonv1.ClusterTaskKind}, // TODO: MAVEN/GRADLE
					Resources: &tektonv1.PipelineTaskResources{
						Inputs: []tektonv1.PipelineTaskInputResource{
							{
								Name:     tmaxv1.PipelineResourceNameGit,
								Resource: tmaxv1.PipelineResourceNameGit,
							},
						},
					},
					Params: []tektonv1.Param{
						{
							Name: tmaxv1.PipelineParamNameSonarUrl,
							Value: tektonv1.ArrayOrString{
								Type:      tektonv1.ParamTypeString,
								StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.PipelineParamNameSonarUrl),
							},
						},
						{
							Name: tmaxv1.PipelineParamNameSonarToken,
							Value: tektonv1.ArrayOrString{
								Type:      tektonv1.ParamTypeString,
								StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.PipelineParamNameSonarToken),
							},
						},
						{
							Name: tmaxv1.PipelineParamNameSonarProjectKey,
							Value: tektonv1.ArrayOrString{
								Type:      tektonv1.ParamTypeString,
								StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.PipelineParamNameSonarProjectKey),
							},
						},
					},
				},
			},
		},
	}
}
