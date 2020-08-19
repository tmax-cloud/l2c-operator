package v1

import (
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/apis/resource/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	"github.com/tmax-cloud/l2c-operator/pkg/sonarqube"
)

func pipelineRun(l2c *tmaxv1.L2c, sonar *sonarqube.SonarQube) *tektonv1.PipelineRun {
	return &tektonv1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      l2c.Name,
			Namespace: l2c.Namespace,
		},
		Spec: tektonv1.PipelineRunSpec{
			PipelineRef:        &tektonv1.PipelineRef{Name: l2c.Name},
			ServiceAccountName: "",
			Resources: []tektonv1.PipelineResourceBinding{
				{
					Name: tmaxv1.PipelineResourceNameGit,
					ResourceSpec: &tektonv1alpha1.PipelineResourceSpec{
						Type: tektonv1alpha1.PipelineResourceTypeGit,
						Params: []v1alpha1.ResourceParam{
							{
								Name:  "url",
								Value: l2c.Spec.Was.From.Git.Url,
							},
						},
					},
				},
			},
			Params: []tektonv1.Param{
				{
					Name: tmaxv1.PipelineParamNameSonarUrl,
					Value: tektonv1.ArrayOrString{
						Type:      tektonv1.ParamTypeString,
						StringVal: sonar.URL,
					},
				},
				{
					Name: tmaxv1.PipelineParamNameSonarToken,
					Value: tektonv1.ArrayOrString{
						Type:      tektonv1.ParamTypeString,
						StringVal: sonar.Token,
					},
				},
				{
					Name: tmaxv1.PipelineParamNameSonarProjectKey,
					Value: tektonv1.ArrayOrString{
						Type:      tektonv1.ParamTypeString,
						StringVal: l2c.GetSonarProjectName(),
					},
				},
			},
		},
	}
}
