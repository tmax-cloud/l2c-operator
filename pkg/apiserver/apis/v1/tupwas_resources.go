package v1

import (
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func tupWasAnalyzePipelineRun(tupWas *tmaxv1.TupWAS) *tektonv1.PipelineRun {
	return &tektonv1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupWas.GenAnalyzePipelineName(),
			Namespace: tupWas.Namespace,
			Labels:    tupWas.GenLabels(),
		},
		Spec: tektonv1.PipelineRunSpec{
			PipelineRef: &tektonv1.PipelineRef{Name: tupWas.GenAnalyzePipelineName()},
			Params: []tektonv1.Param{{
				Name:  tmaxv1.WasPipelineParamNameProjectId,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupWas.Name},
			}, {
				Name:  tmaxv1.WasPipelineParamNameGitUrl,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupWas.Spec.From.Git.Url},
			}, {
				Name:  tmaxv1.WasPipelineParamNameGitRev,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupWas.Spec.From.Git.Revision},
			}},
			Workspaces: []tektonv1.WorkspaceBinding{{
				Name:                  tmaxv1.WasPipelineWorkspaceName,
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: tupWas.GenResourceName()},
			}},
		},
	}
}

func tupWasBuildDeployPipelineRun(tupWas *tmaxv1.TupWAS) *tektonv1.PipelineRun {
	return &tektonv1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupWas.GenBuildDeployPipelineName(),
			Namespace: tupWas.Namespace,
			Labels:    tupWas.GenLabels(),
		},
		Spec: tektonv1.PipelineRunSpec{
			PipelineRef:        &tektonv1.PipelineRef{Name: tupWas.GenBuildDeployPipelineName()},
			ServiceAccountName: tupWas.GenResourceName(),
			Params: []tektonv1.Param{{
				Name:  tmaxv1.WasPipelineParamNameAppName,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupWas.Name},
			}, {
				Name:  tmaxv1.WasPipelineParamNameDeployCfg,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupWas.GenWasResourceName()},
			}},
			Workspaces: []tektonv1.WorkspaceBinding{{
				Name:                  tmaxv1.WasPipelineWorkspaceName,
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: tupWas.GenResourceName()},
				SubPath:               "project/" + tupWas.Name,
			}},
		},
	}
}
