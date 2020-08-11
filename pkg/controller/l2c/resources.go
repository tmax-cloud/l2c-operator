package l2c

import (
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func pipeline(name, ns string) *tektonv1.Pipeline {
	return &tektonv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: tektonv1.PipelineSpec{},
	}
}
