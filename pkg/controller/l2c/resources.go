package l2c

import (
	"fmt"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

func configMap(l2c *tmaxv1.L2c) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      l2c.Name,
			Namespace: l2c.Namespace,
		},
		Data: map[string]string{},
	}
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
					TaskRef: &tektonv1.TaskRef{Name: tmaxv1.TaskNameAnalyzeMaven, Kind: tektonv1.ClusterTaskKind}, //TODO: MAVEN/GRADLE
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
