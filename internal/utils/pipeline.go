package utils

import (
	"fmt"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

func GetTaskRunStatus(pr *tektonv1.PipelineRun, name tmaxv1.PipelineTaskName) (*tektonv1.PipelineRunTaskRunStatus, error) {
	for _, v := range pr.Status.TaskRuns {
		if v.PipelineTaskName == string(name) {
			return v, nil
		}
	}

	return nil, fmt.Errorf("no task %s in pipelineRun %s", string(name), pr.Name)
}
