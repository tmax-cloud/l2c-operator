package l2c

import (
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

func pipeline(l2c *tmaxv1.L2c) *tektonv1.Pipeline {
	return &tektonv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      l2c.Name,
			Namespace: l2c.Namespace,
		},
		Spec: tektonv1.PipelineSpec{
			Description: "dummy pipeline",
		},
	}
}

func pipelineRun(l2c *tmaxv1.L2c) *tektonv1.PipelineRun {
	return &tektonv1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: l2c.Name,
			Namespace: l2c.Namespace,
		},
		Spec: tektonv1.PipelineRunSpec{

		},
	}
}
