package v1

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/operator-framework/operator-sdk/pkg/status"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tmax-cloud/l2c-operator/internal/utils"
	"github.com/tmax-cloud/l2c-operator/internal/wrapper"
	"github.com/tmax-cloud/l2c-operator/pkg/apis"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ApiType string

const (
	ApiTypeAnalyze = ApiType("analyze")
	ApiTypeRun     = ApiType("run")
)

func AddTupWasApis(parent *wrapper.RouterWrapper) error {
	tupWasWrapper := wrapper.New(fmt.Sprintf("/%s/{tupName}", TupWasKind), nil, nil)
	if err := parent.Add(tupWasWrapper); err != nil {
		return err
	}

	tupWasWrapper.Router.Use(Authorize)

	if err := addTupWasAnalyzeApi(tupWasWrapper); err != nil {
		return err
	}
	if err := addTupWasRunApi(tupWasWrapper); err != nil {
		return err
	}
	return nil
}

func addTupWasAnalyzeApi(parent *wrapper.RouterWrapper) error {
	runWrapper := wrapper.New("/analyze", []string{"PUT"}, tupWasAnalyzeHandler)
	if err := parent.Add(runWrapper); err != nil {
		return err
	}

	return nil
}

func addTupWasRunApi(parent *wrapper.RouterWrapper) error {
	runWrapper := wrapper.New("/run", []string{"PUT"}, tupWasRunHandler)
	if err := parent.Add(runWrapper); err != nil {
		return err
	}

	return nil
}

func tupWasAnalyzeHandler(w http.ResponseWriter, req *http.Request) {
	tupWasApiHandler(w, req, ApiTypeAnalyze)
}

func tupWasRunHandler(w http.ResponseWriter, req *http.Request) {
	tupWasApiHandler(w, req, ApiTypeRun)
}

func tupWasApiHandler(w http.ResponseWriter, req *http.Request, apiType ApiType) {
	vars := mux.Vars(req)

	ns, nsExist := vars["namespace"]
	resourceName, nameExist := vars["tupName"]
	if !nsExist || !nameExist {
		_ = utils.RespondError(w, http.StatusBadRequest, "url is malformed")
		return
	}

	opt := client.Options{}
	utils.AddSchemes(&opt, schema.GroupVersion{Group: "tmax.io", Version: "v1"}, &tmaxv1.TupWAS{})
	if err := tektonv1.AddToScheme(opt.Scheme); err != nil {
		log.Error(err, "")
		_ = utils.RespondError(w, http.StatusInternalServerError, "could not initialize client")
		return
	}

	c, err := utils.Client(opt)
	if err != nil {
		log.Error(err, "cannot get client")
		_ = utils.RespondError(w, http.StatusInternalServerError, "could not make k8s client")
		return
	}

	tupWas := &tmaxv1.TupWAS{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: resourceName, Namespace: ns}, tupWas); err != nil {
		log.Error(err, "cannot get tupWas")
		if errors.IsNotFound(err) {
			_ = utils.RespondError(w, http.StatusNotFound, fmt.Sprintf("there is no TupWAS %s/%s", ns, resourceName))
		} else {
			_ = utils.RespondError(w, http.StatusInternalServerError, "cannot get tupWas")
		}
		return
	}

	// Declare variables depending on apiType
	var cond *status.Condition
	var condFound bool
	var pr *tektonv1.PipelineRun
	var msg string
	switch apiType {
	case ApiTypeAnalyze:
		cond, condFound = tupWas.Status.GetCondition(tmaxv1.WasConditionKeyProjectAnalyzing)
		pr = tupWasAnalyzePipelineRun(tupWas)
		msg = fmt.Sprintf("tupWas %s has started analyzing", tupWas.Name)

		// Check if TupWAS project is ready, if not, return error
		readyCond, ok := tupWas.Status.GetCondition(tmaxv1.WasConditionKeyProjectReady)
		if !ok || readyCond.Status != corev1.ConditionTrue {
			_ = utils.RespondError(w, http.StatusAccepted, "TupWAS is not ready yet")
			return
		}
	case ApiTypeRun:
		cond, condFound = tupWas.Status.GetCondition(tmaxv1.WasConditionKeyProjectRunning)
		pr = tupWasBuildDeployPipelineRun(tupWas)
		msg = fmt.Sprintf("tupWas %s has started running", tupWas.Name)

		// Check if analyze is complete
		if tupWas.Status.LastAnalyzeResult != string(tektonv1.TaskRunReasonSuccessful) {
			_ = utils.RespondError(w, http.StatusAccepted, "TupWAS is not analyzed successfully yet")
			return
		}
	default:
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("api type %s is not supported", string(apiType)))
		return
	}

	if !condFound || cond == nil || pr == nil {
		_ = utils.RespondError(w, http.StatusAccepted, "TupWAS may not be ready yet")
		return
	}

	// Check if status is analyzing|running
	if cond.Status == corev1.ConditionTrue {
		_ = utils.RespondError(w, http.StatusAccepted, fmt.Sprintf("TupWAS process is still in condition %s", string(apiType)))
		return
	}

	// Now, we can create PR
	s := runtime.NewScheme()
	if err := apis.AddToScheme(s); err != nil {
		log.Error(err, "")
		_ = utils.RespondError(w, http.StatusInternalServerError, "cannot make new scheme")
		return
	}
	if err := utils.CheckAndCreateObject(pr, tupWas, c, s, true); err != nil {
		_ = utils.RespondError(w, http.StatusAccepted, "cannot create PipelineRun")
		return
	}

	_ = utils.RespondJSON(w, map[string]string{"message": msg})
	log.Info(fmt.Sprintf("Created pipelineRun %s/%s", pr.Namespace, pr.Name))
}
