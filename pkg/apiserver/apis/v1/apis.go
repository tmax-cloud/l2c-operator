package v1

import (
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	"github.com/tmax-cloud/l2c-operator/internal/wrapper"
)

const (
	ApiGroup   = "tup.tmax.io"
	ApiVersion = "v1"
	TupDbKind  = "tupdbs"
	TupWasKind = "tupwas"
)

var log = logf.Log.WithName("l2c-apis")

func AddV1Apis(parent *wrapper.RouterWrapper) error {
	versionWrapper := wrapper.New(fmt.Sprintf("/%s/%s", ApiGroup, ApiVersion), nil, versionHandler)
	if err := parent.Add(versionWrapper); err != nil {
		return err
	}

	namespaceWrapper := wrapper.New("/namespaces/{namespace}", nil, nil)
	if err := versionWrapper.Add(namespaceWrapper); err != nil {
		return err
	}

	if err := AddTupWasApis(namespaceWrapper); err != nil {
		return err
	}

	if err := AddTupDBApis(namespaceWrapper); err != nil {
		return err
	}

	return nil
}

func versionHandler(w http.ResponseWriter, _ *http.Request) {
	apiResourceList := &metav1.APIResourceList{}
	apiResourceList.Kind = "APIResourceList"
	apiResourceList.GroupVersion = fmt.Sprintf("%s/%s", ApiGroup, ApiVersion)
	apiResourceList.APIVersion = ApiVersion

	apiResourceList.APIResources = []metav1.APIResource{
		{
			Name:       fmt.Sprintf("%s/analyze", TupWasKind),
			Namespaced: true,
		},
		{
			Name:       fmt.Sprintf("%s/run", TupWasKind),
			Namespaced: true,
		},
		{
			Name:       fmt.Sprintf("%s/analyze", TupDbKind),
			Namespaced: true,
		},
		{
			Name:       fmt.Sprintf("%s/migrate", TupDbKind),
			Namespaced: true,
		},
	}

	_ = utils.RespondJSON(w, apiResourceList)
}
