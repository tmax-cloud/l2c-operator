package v1

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"net/http"

	"github.com/gorilla/mux"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	"github.com/tmax-cloud/l2c-operator/internal/wrapper"
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

	log.Info("RUN!")
	// TODO : actual logic to run

	_ = utils.RespondJSON(w, l2c)
}
