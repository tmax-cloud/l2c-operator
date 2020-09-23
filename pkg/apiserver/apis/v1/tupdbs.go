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
	tupdbcontroller "github.com/tmax-cloud/l2c-operator/pkg/controller/tupdb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TupDBApiTypeAnalyze = ApiType("analyze")
	TupDBApiTypeMigrate = ApiType("migrate")
)

func AddTupDBApis(parent *wrapper.RouterWrapper) error {
	tupDBWrapper := wrapper.New(fmt.Sprintf("/%s/{tupName}", TupDbKind), nil, nil)
	if err := parent.Add(tupDBWrapper); err != nil {
		return err
	}

	tupDBWrapper.Router.Use(Authorize)

	if err := addTupDBAnalyzeApi(tupDBWrapper); err != nil {
		return err
	}

	if err := addTupDBMigrateApi(tupDBWrapper); err != nil {
		return err
	}

	return nil
}

func addTupDBAnalyzeApi(parent *wrapper.RouterWrapper) error {
	runWrapper := wrapper.New("/analyze", []string{"PUT"}, tupDBAnalyzeHandler)
	if err := parent.Add(runWrapper); err != nil {
		return err
	}

	return nil
}

func addTupDBMigrateApi(parent *wrapper.RouterWrapper) error {
	runWrapper := wrapper.New("/migrate", []string{"PUT"}, tupDBMigrateHandler)
	if err := parent.Add(runWrapper); err != nil {
		return err
	}
	return nil
}

func tupDBAnalyzeHandler(w http.ResponseWriter, req *http.Request) {
	tupDBApiHandler(w, req, TupDBApiTypeAnalyze)
}

func tupDBMigrateHandler(w http.ResponseWriter, req *http.Request) {
	tupDBApiHandler(w, req, TupDBApiTypeMigrate)
}

func tupDBApiHandler(w http.ResponseWriter, req *http.Request, apiType ApiType) {
	vars := mux.Vars(req)

	namespace, namespaceExist := vars["namespace"]
	tupDBName, nameExist := vars["tupName"]
	if !namespaceExist || !nameExist {
		_ = utils.RespondError(w, http.StatusBadRequest, "url is malformed")
		return
	}
	logger := utils.GetTupLogger(tmaxv1.TupDB{}, namespace, tupDBName)
	logger.Info("Api Handler came", "Type", apiType)

	opt := client.Options{}
	utils.AddSchemes(&opt, schema.GroupVersion{Group: "tmax.io", Version: "v1"}, &tmaxv1.TupDB{})
	if err := tektonv1.AddToScheme(opt.Scheme); err != nil {
		log.Error(err, "Add scheme error")
		_ = utils.RespondError(w, http.StatusInternalServerError, "could not initialize client")
		return
	}

	c, err := utils.Client(opt)
	if err != nil {
		log.Error(err, "cannot get client")
		_ = utils.RespondError(w, http.StatusInternalServerError, "could not make k8s client")
		return
	}

	tupDB := &tmaxv1.TupDB{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: tupDBName, Namespace: namespace}, tupDB); err != nil {
		logger.Error(err, "cannot get tupDB")
		if errors.IsNotFound(err) {
			_ = utils.RespondError(w, http.StatusNotFound, fmt.Sprintf("There is no TupDB %s/%s", namespace, tupDBName))
		} else {
			_ = utils.RespondError(w, http.StatusInternalServerError, "cannot get tupDB")
		}
		return
	}

	var cond *status.Condition
	var condFound bool
	var pipelineRun *tektonv1.PipelineRun
	var msg string
	switch apiType {
	case TupDBApiTypeAnalyze:
		cond, condFound = tupDB.Status.GetCondition(tmaxv1.DBConditionKeyDBAnalyzing)
		//pipelineRun = AnalyzeDBPipelineRun(tupDB)
		msg = fmt.Sprintf("tupDB %s has started anaylzing", tupDB.Name)

	case TupDBApiTypeMigrate:
		cond, condFound = tupDB.Status.GetCondition(tmaxv1.DBConditionKeyDBMigrating)
		pipelineRun = tupdbcontroller.MigratePipelineRun(tupDB)
		msg = fmt.Sprintf("tupDB %s has started running", tupDB.Name)
		// [TODO] DB deploy should be first task
		// [TODO] Analyze Result check
	default:
		_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Api type %s is not supported", string(apiType)))
		return
	}
	if !condFound || cond == nil || pipelineRun == nil {
		_ = utils.RespondError(w, http.StatusAccepted, "TupDB may not be ready yet")
		return
	}

	if cond.Status == corev1.ConditionTrue {
		_ = utils.RespondError(w, http.StatusAccepted, fmt.Sprintf("TupDB process is still in condtion %s", string(apiType)))
	}

	s := runtime.NewScheme()
	if err := apis.AddToScheme(s); err != nil {
		log.Error(err, "")
		_ = utils.RespondError(w, http.StatusInternalServerError, "cannot make new scheme")
		return
	}

	if err := utils.CheckAndCreateObject(pipelineRun, tupDB, c, s, true); err != nil {
		_ = utils.RespondError(w, http.StatusAccepted, "cannot create PipelineRun")
		return
	}

	_ = utils.RespondJSON(w, map[string]string{"message": msg})
	log.Info(fmt.Sprintf("Created pipelineRun %s/%s", pipelineRun.Namespace, pipelineRun.Name))
}
