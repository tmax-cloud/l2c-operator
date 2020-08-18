package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	"github.com/tmax-cloud/l2c-operator/internal/wrapper"
	"github.com/tmax-cloud/l2c-operator/pkg/apis"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

func AddRunApis(parent *wrapper.RouterWrapper) error {
	l2cWrapper := wrapper.New(fmt.Sprintf("/%s/{l2cName}", L2cKind), nil, nil)
	if err := parent.Add(l2cWrapper); err != nil {
		return err
	}

	l2cWrapper.Router.Use(Authorize)

	if err := addRunApi(l2cWrapper); err != nil {
		return err
	}
	return nil
}

func addRunApi(parent *wrapper.RouterWrapper) error {
	runWrapper := wrapper.New("/run", []string{"PUT"}, runHandler)
	if err := parent.Add(runWrapper); err != nil {
		return err
	}

	return nil
}

func runHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	ns, nsExist := vars["namespace"]
	l2cName, nameExist := vars["l2cName"]
	if !nsExist || !nameExist {
		_ = utils.RespondError(w, http.StatusBadRequest, "url is malformed")
		return
	}

	opt := client.Options{}
	utils.AddSchemes(&opt, schema.GroupVersion{Group: "tmax.io", Version: "v1"}, &tmaxv1.L2c{})
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

	l2c := &tmaxv1.L2c{}
	if err := c.Get(context.TODO(), types.NamespacedName{Name: l2cName, Namespace: ns}, l2c); err != nil {
		log.Error(err, "cannot get l2c")
		if errors.IsNotFound(err) {
			_ = utils.RespondError(w, http.StatusNotFound, fmt.Sprintf("there is no L2c %s/%s", ns, l2cName))
		} else {
			_ = utils.RespondError(w, http.StatusInternalServerError, "cannot get l2c")
		}
		return
	}

	// Check if L2c project is ready, if not, return error
	readyCond, ok := l2c.Status.GetCondition(tmaxv1.ConditionKeyProjectReady)
	if !ok || readyCond.Status != corev1.ConditionTrue {
		_ = utils.RespondError(w, http.StatusAccepted, "L2c is not ready yet")
		return
	}

	// Check if L2c is currently running
	runningCond, rcFound := l2c.Status.GetCondition(tmaxv1.ConditionKeyProjectRunning)
	if rcFound {
		if runningCond.Status == corev1.ConditionTrue {
			_ = utils.RespondError(w, http.StatusAccepted, "L2c process is still running")
			return
		}
	} else {
		_ = utils.RespondError(w, http.StatusAccepted, "L2c may not be ready yet")
		return
	}

	pr := pipelineRun(l2c, sonar)
	// Delete first
	if err := c.Delete(context.TODO(), pr); err != nil && !errors.IsNotFound(err) {
		log.Error(err, "")
		_ = utils.RespondError(w, http.StatusInternalServerError, "cannot delete existing PipelineRun")
		return
	}

	// Now, we can create PR
	pr = pipelineRun(l2c, sonar) // Intact pipelineRun for creation
	s := runtime.NewScheme()
	if err := apis.AddToScheme(s); err != nil {
		log.Error(err, "")
		_ = utils.RespondError(w, http.StatusInternalServerError, "cannot make new scheme")
		return
	}
	if err := controllerutil.SetControllerReference(l2c, pr, s); err != nil {
		log.Error(err, "")
		_ = utils.RespondError(w, http.StatusInternalServerError, "cannot set ownerReference to PipelineRun")
		return
	}
	if err := c.Create(context.TODO(), pr); err != nil {
		log.Error(err, "")
		_ = utils.RespondError(w, http.StatusInternalServerError, "cannot create PipelineRun")
		return
	}

	_ = utils.RespondJSON(w, map[string]string{"message": fmt.Sprintf("l2c %s has started", l2c.Name)})
	log.Info(fmt.Sprintf("Created pipelineRun %s/%s", pr.Namespace, pr.Name))
}
