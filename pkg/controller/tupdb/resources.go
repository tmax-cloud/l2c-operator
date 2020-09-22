package tupdb

import (
	"fmt"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)


func analyzePipeline(tupDB *tmaxv1.TupDB) *tektonv1.Pipeline {
	return &tektonv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupDB.GenAnalyzePipelineName(),
			Namespace: tupDB.Namespace,
			Labels:    tupDB.GenLabels(),
		},
		Spec: tektonv1.PipelineSpec{
			Params: []tektonv1.ParamSpec{
				{Name: tmaxv1.DBAnalyzePipelineParamTarget},
				{Name: tmaxv1.DBAnalyzePipelineParamFileType},
				{Name: tmaxv1.DBAnalyzePipelineParamFileSyntax},
				{Name: tmaxv1.DBAnalyzePipelineParamFileExtension},
				{Name: tmaxv1.DBAnalyzePipelineParamFileSearch},
				{Name: tmaxv1.DBAnalyzePipelineParamFileLocation},
				{Name: tmaxv1.DBAnalyzePipelineParamFileCharset},
				{Name: tmaxv1.DBAnalyzePipelineParamReportOptions},
			},
			Tasks: []tektonv1.PipelineTask{{
				Name:    tmaxv1.DBPipelineTaskNameAnalyzeDB,
				TaskRef: &tektonv1.TaskRef{Name: tmaxv1.TaskNameAnalyzeDB, Kind: tektonv1.ClusterTaskKind},
				Params: []tektonv1.Param{{
					Name:  "target",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBAnalyzePipelineParamTarget)},
				}, {
					Name:  "fileType",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBAnalyzePipelineParamFileType)},
				}, {
					Name:  "fileSyntax",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBAnalyzePipelineParamFileSyntax)},
				}, {
					Name:  "fileExtension",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBAnalyzePipelineParamFileExtension)},
				}, {
					Name:  "fileSearch",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBAnalyzePipelineParamFileSearch)},
				}, {
					Name:  "fileLocation",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBAnalyzePipelineParamFileLocation)},
				}, {
					Name:  "fileCharset",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBAnalyzePipelineParamFileCharset)},
				}, {
					Name:  "reportOptions",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBAnalyzePipelineParamReportOptions)},
				}},
			}},
		},
	}
}

func MigratePipeline(tupDB *tmaxv1.TupDB) *tektonv1.Pipeline {
	return &tektonv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupDB.GenMigratePipelineName(),
			Namespace: tupDB.Namespace,
			Labels:    tupDB.GenLabels(),
		},
		Spec: tektonv1.PipelineSpec{
			Params: []tektonv1.ParamSpec{
				{Name: tmaxv1.DBPipelineParamNameSourceUserName},
				{Name: tmaxv1.DBPipelineParamNameSourcePassword},
				{Name: tmaxv1.DBPipelineParamNameSourceType},
				{Name: tmaxv1.DBPipelineParamNameSourceSID},
				{Name: tmaxv1.DBPipelineParamNameSourceAs},
				{Name: tmaxv1.DBPipelineParamNameSourcePort},
				{Name: tmaxv1.DBPipelineParamNameSourceIP},
				{Name: tmaxv1.DBPipelineParamNameTargetUserName},
				{Name: tmaxv1.DBPipelineParamNameTargetPassword},
				{Name: tmaxv1.DBPipelineParamNameTargetType},
				{Name: tmaxv1.DBPipelineParamNameTargetSID},
				{Name: tmaxv1.DBPipelineParamNameTargetPort},
				{Name: tmaxv1.DBPipelineParamNameTargetIP},
				{Name: tmaxv1.DBPipelineParamNameTargetUser},
				{Name: tmaxv1.DBPipelineParamNameFull},
			},
			Tasks: []tektonv1.PipelineTask {{
				Name: tmaxv1.DBPipelineTaskNameMigrateDB,
				TaskRef: &tektonv1.TaskRef{Name: tmaxv1.TaskNameMigrateDB, Kind: tektonv1.ClusterTaskKind},
				Params: []tektonv1.Param{{
					// [TODO] Parameterize
					Name: "SECRET_NAME",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: "tup-db-secret"},
				}, {
					Name: "SOURCE_TYPE",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBPipelineParamNameSourceType)},
				}, {
					Name: "SOURCE_PORT",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBPipelineParamNameSourcePort)},
				}, {
					Name: "SOURCE_HOST",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBPipelineParamNameSourceIP)},
				}, {
					Name: "TARGET_TYPE",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBPipelineParamNameTargetType)},
				}, {
					Name: "TARGET_PORT",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBPipelineParamNameTargetPort)},
				}, {
					Name: "TARGET_HOST",
					Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: fmt.Sprintf("$(params.%s)", tmaxv1.DBPipelineParamNameTargetIP)},
				}},
			}},
		},
	}
}

func AnalyzePipelineRun(tupDB *tmaxv1.TupDB) *tektonv1.PipelineRun {
	return &tektonv1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupDB.GenAnalyzePipelineName(),
			Namespace: tupDB.Namespace,
			Labels:    tupDB.GenLabels(),
		},
		Spec: tektonv1.PipelineRunSpec{
			PipelineRef: &tektonv1.PipelineRef{Name: tupDB.GenAnalyzePipelineName()},
			// [TODO]
			Params: []tektonv1.Param{{}},
		},
	}
}

func MigratePipelineRun(tupDB *tmaxv1.TupDB) *tektonv1.PipelineRun {
	return &tektonv1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tupDB.GenMigratePipelineName(),
			Namespace: tupDB.Namespace,
			Labels:    tupDB.GenLabels(),
		},
		Spec: tektonv1.PipelineRunSpec{
			PipelineRef: &tektonv1.PipelineRef{Name: tupDB.GenMigratePipelineName()},
			// [TODO]
			Params: []tektonv1.Param{{
				Name: tmaxv1.DBPipelineParamNameSourceUserName,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupDB.Spec.From.User},
			}, {
				Name: tmaxv1.DBPipelineParamNameSourcePassword,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupDB.Spec.From.Password},
			}, {
				Name: tmaxv1.DBPipelineParamNameSourceType,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupDB.Spec.From.Type},
			}, {
				Name: tmaxv1.DBPipelineParamNameSourceSID,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupDB.Spec.From.Sid},
			}, {
				Name: tmaxv1.DBPipelineParamNameSourceAs,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: "NORMAL"},
			}, {
				Name: tmaxv1.DBPipelineParamNameSourcePort,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: strconv.Itoa(int(tupDB.Spec.From.Port))},
			}, {
				Name: tmaxv1.DBPipelineParamNameSourceIP,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupDB.Spec.From.Host},
			}, {
				Name: tmaxv1.DBPipelineParamNameTargetUserName,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupDB.Spec.To.User},
			}, {
				Name: tmaxv1.DBPipelineParamNameTargetPassword,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupDB.Spec.To.Password},
			}, {
				Name: tmaxv1.DBPipelineParamNameTargetType,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupDB.Spec.To.Type},
			}, {
				Name: tmaxv1.DBPipelineParamNameTargetSID,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupDB.Spec.To.Sid},
			}, {
				Name: tmaxv1.DBPipelineParamNameTargetPort,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: strconv.Itoa(int(tupDB.Status.TargetPort))},
			}, {
				Name: tmaxv1.DBPipelineParamNameTargetIP,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupDB.Status.TargetHost},
			}, {
				Name: tmaxv1.DBPipelineParamNameTargetUser,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: tupDB.Spec.To.User}, // [TODO] User == Username?
			}, {
				Name: tmaxv1.DBPipelineParamNameFull,
				Value: tektonv1.ArrayOrString{Type: tektonv1.ParamTypeString, StringVal: "YES"}, // [TODO] Full Configuration?
			}},
		},
	}
}
