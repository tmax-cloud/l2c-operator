package sonarqube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	"github.com/tmax-cloud/l2c-operator/pkg/apis"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

func (s *SonarServer) WebhookHandleFunc(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer req.Body.Close()

	data := &tmaxv1.SonarWebhookRequest{}
	if err := json.Unmarshal(body, data); err != nil {
		log.Error(err, "unable to unmarshal json")
		_ = utils.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get namespace/name of l2c by splitting sonar project key
	values := strings.Split(data.Project.Key, "_")
	if len(values) != 2 {
		msg := "project is not created from l2c operator"
		log.Error(fmt.Errorf(msg), "")
		_ = utils.RespondError(w, http.StatusBadRequest, msg)
		return
	}

	namespace := values[0]
	name := values[1]

	l2c := &tmaxv1.L2c{}
	if err := s.c.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, l2c); err != nil {
		log.Error(err, "cannot get l2c object")
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Check if pipelineRun is running & Analyze task is running
	prName := l2c.Status.PipelineRunName
	if prName == "" {
		msg := "PipelineRun is not running but webhook arrived"
		log.Error(fmt.Errorf(msg), "")
		_ = utils.RespondError(w, http.StatusInternalServerError, msg)
		return
	}

	pr := &tektonv1.PipelineRun{}
	if err := s.c.Get(context.TODO(), types.NamespacedName{Name: prName, Namespace: namespace}, pr); err != nil {
		log.Error(err, "cannot get pipelineRun object")
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	trStatus, err := utils.GetTaskRunStatus(pr, tmaxv1.PipelineTaskNameAnalyze)
	if err != nil {
		log.Error(err, "cannot get TaskRun status")
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(trStatus.Status.Conditions) != 1 || trStatus.Status.Conditions[0].Status != corev1.ConditionUnknown || trStatus.Status.Conditions[0].Reason != "Running" {
		log.Error(err, "analyze task is not currently running")
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Here, analyze task is running... get Pod
	podName := trStatus.Status.PodName
	if podName == "" {
		msg := "pod name is invalid"
		log.Error(fmt.Errorf(msg), "")
		_ = utils.RespondError(w, http.StatusInternalServerError, msg)
		return
	}
	pod := &corev1.Pod{}
	if err := s.c.Get(context.TODO(), types.NamespacedName{Name: podName, Namespace: namespace}, pod); err != nil {
		log.Error(err, "cannot get taskRun pod")
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Do something, depending on the analysis result
	var result apis.ScanResult
	if data.QualityGate.Status == "OK" {
		result = apis.ScanResultOk
	} else {
		result = apis.ScanResultFail
	}
	if err := s.execPodCommand(pod, result); err != nil {
		log.Error(err, "cannot exec cmd to a pod")
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *SonarServer) execPodCommand(pod *corev1.Pod, result apis.ScanResult) error {
	command := []string{"/scan-waiter", string(result)}

	// Create custom REST API call to exec
	req := s.clientSet.CoreV1().RESTClient().Post().Resource("pods").Name(pod.Name).Namespace(pod.Namespace).SubResource("exec")
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return err
	}

	parameterCode := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&corev1.PodExecOptions{
		Command:   command,
		Container: fmt.Sprintf("step-%s", apis.WaiterContainerName),
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, parameterCode)

	exec, err := remotecommand.NewSPDYExecutor(s.clientCfg, "POST", req.URL())
	if err != nil {
		return err
	}

	stdOutBuf := &bytes.Buffer{}
	stdErrBuf := &bytes.Buffer{}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: stdOutBuf,
		Stderr: stdErrBuf,
	})
	if err != nil {
		return err
	}

	errString := stdErrBuf.String()
	if errString != "" {
		return fmt.Errorf(errString)
	}

	return nil
}
