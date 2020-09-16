package tupwas

import (
	"bytes"
	"fmt"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/tmax-cloud/l2c-operator/internal"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

// ConfigMap for WAS deployment spec
func wasDeployConfigMap(tupWas *tmaxv1.TupWAS) (*corev1.ConfigMap, error) {
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
		Yaml:   true,
		Pretty: true,
		Strict: true,
	})

	// Deployment object
	deploy, err := wasDeploy(tupWas)
	if err != nil {
		return nil, err
	}
	deployBuf := new(bytes.Buffer)
	if err := serializer.Encode(deploy, deployBuf); err != nil {
		return nil, err
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupWas.GenWasResourceName(),
			Namespace: tupWas.Namespace,
			Labels:    tupWas.GenLabels(),
		},
		Data: map[string]string{
			"deploy-spec.yaml": deployBuf.String(),
		},
	}, nil
}

func wasDeployServiceAccount(tupWas *tmaxv1.TupWAS) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupWas.GenResourceName(),
			Namespace: tupWas.Namespace,
			Labels:    tupWas.GenLabels(),
		},
	}
}

func wasDeployRoleBinding(tupWas *tmaxv1.TupWAS) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupWas.GenResourceName(),
			Namespace: tupWas.Namespace,
			Labels:    tupWas.GenLabels(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "l2c",
		},
		Subjects: []rbacv1.Subject{{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      tupWas.GenResourceName(),
			Namespace: tupWas.Namespace,
		}},
	}
}

func gitReportPVC(tupWas *tmaxv1.TupWAS) (*corev1.PersistentVolumeClaim, error) {
	storageQuantity, err := resource.ParseQuantity(internal.WasProjectStorageSize)
	if err != nil {
		return nil, err
	}
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupWas.GenResourceName(),
			Namespace: tupWas.Namespace,
			Labels:    tupWas.GenLabels(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &internal.StorageClassName,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": storageQuantity,
				},
			},
		},
	}, nil
}

func analyzePipeline(tupWas *tmaxv1.TupWAS) *tektonv1.Pipeline {
	return &tektonv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupWas.GenAnalyzePipelineName(),
			Namespace: tupWas.Namespace,
			Labels:    tupWas.GenLabels(),
		},
		Spec: tektonv1.PipelineSpec{
			Params: []tektonv1.ParamSpec{
				{Name: tmaxv1.WasPipelineParamNameProjectId},
				{Name: tmaxv1.WasPipelineParamNameGitUrl},
				{
					Name:    tmaxv1.WasPipelineParamNameGitRev,
					Default: &tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: "master"},
				},
				{Name: tmaxv1.WasPipelineParamNameSourceType},
				{Name: tmaxv1.WasPipelineParamNameTargetType},
			},
			Workspaces: []tektonv1.PipelineWorkspaceDeclaration{{Name: tmaxv1.WasPipelineWorkspaceName}},
			Tasks: []tektonv1.PipelineTask{{
				Name:    string(tmaxv1.WasPipelineTaskNameClone),
				TaskRef: &tektonv1.TaskRef{Name: tmaxv1.TaskNameGitClone, Kind: tektonv1.ClusterTaskKind},
				Params: []tektonv1.Param{{
					Name:  "skipIfExists",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: "true"},
				}, {
					Name:  "url",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.WasPipelineParamNameGitUrl)},
				}, {
					Name:  "revision",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.WasPipelineParamNameGitRev)},
				}, {
					Name:  "subdirectory",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.WasPipelineParamNameProjectId)},
				}},
				Workspaces: []tektonv1.WorkspacePipelineTaskBinding{{
					Name:      "output",
					Workspace: tmaxv1.WasPipelineWorkspaceName,
					SubPath:   "project",
				}},
			}, {
				Name:     string(tmaxv1.WasPipelineTaskNameAnalyze),
				TaskRef:  &tektonv1.TaskRef{Name: tmaxv1.TaskNameAnalyzeWas, Kind: tektonv1.ClusterTaskKind},
				RunAfter: []string{string(tmaxv1.WasPipelineTaskNameClone)},
				Params: []tektonv1.Param{{
					Name:  "project-id",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.WasPipelineParamNameProjectId)},
				}, {
					Name:  "source-type",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.WasPipelineParamNameSourceType)},
				}, {
					Name:  "target-type",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.WasPipelineParamNameTargetType)},
				}},
				Workspaces: []tektonv1.WorkspacePipelineTaskBinding{{
					Name:      "source",
					Workspace: tmaxv1.WasPipelineWorkspaceName,
					SubPath:   "project",
				}, {
					Name:      "report",
					Workspace: tmaxv1.WasPipelineWorkspaceName,
					SubPath:   "report",
				}},
			}},
		},
	}
}

func buildDeployPipeline(tupWas *tmaxv1.TupWAS) (*tektonv1.Pipeline, error) {
	builderImg, err := tupWas.GenBuilderImage()
	if err != nil {
		return nil, err
	}
	return &tektonv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupWas.GenBuildDeployPipelineName(),
			Namespace: tupWas.Namespace,
			Labels:    tupWas.GenLabels(),
		},
		Spec: tektonv1.PipelineSpec{
			Params: []tektonv1.ParamSpec{
				{Name: tmaxv1.WasPipelineParamNameAppName},
				{Name: tmaxv1.WasPipelineParamNameDeployCfg, Default: &tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: ""}},
			},
			Workspaces: []tektonv1.PipelineWorkspaceDeclaration{{Name: tmaxv1.WasPipelineWorkspaceName}},
			Tasks: []tektonv1.PipelineTask{{
				Name:    string(tmaxv1.WasPipelineTaskNameBuild),
				TaskRef: &tektonv1.TaskRef{Name: tmaxv1.TaskNameBuild, Kind: tektonv1.ClusterTaskKind},
				Params: []tektonv1.Param{{
					Name:  "BUILDER_IMAGE",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: builderImg},
				}, {
					Name:  "IMAGE_URL",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupWas.Spec.To.Image.Url},
				}},
				Workspaces: []tektonv1.WorkspacePipelineTaskBinding{{
					Name:      "git-source",
					Workspace: tmaxv1.WasPipelineWorkspaceName,
				}},
			}, {
				Name:     string(tmaxv1.WasPipelineTaskNameDeploy),
				TaskRef:  &tektonv1.TaskRef{Name: tmaxv1.TaskNameDeploy, Kind: tektonv1.ClusterTaskKind},
				RunAfter: []string{string(tmaxv1.WasPipelineTaskNameBuild)},
				Params: []tektonv1.Param{
					{Name: "app-name", Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.WasPipelineParamNameAppName)}},
					{Name: "image-url", Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(tasks.%s.results.image-url)", tmaxv1.WasPipelineTaskNameBuild)}},
					{Name: "deploy-cfg-name", Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.WasPipelineParamNameDeployCfg)}},
					{Name: "deploy-env-json", Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: "{}"}},
				},
			}},
		},
	}, nil
}

func AnalyzePipelineRun(tupWas *tmaxv1.TupWAS) *tektonv1.PipelineRun {
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
			}, {
				Name:  tmaxv1.WasPipelineParamNameSourceType,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupWas.Spec.From.Type},
			}, {
				Name:  tmaxv1.WasPipelineParamNameTargetType,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupWas.Spec.To.Type},
			}},
			Workspaces: []tektonv1.WorkspaceBinding{{
				Name:                  tmaxv1.WasPipelineWorkspaceName,
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: tupWas.GenResourceName()},
			}},
		},
	}
}

func BuildDeployPipelineRun(tupWas *tmaxv1.TupWAS) *tektonv1.PipelineRun {
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
