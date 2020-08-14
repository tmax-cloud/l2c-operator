package v1

import (
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	"github.com/tmax-cloud/l2c-operator/internal/wrapper"
	"github.com/tmax-cloud/l2c-operator/pkg/sonarqube"
)

const (
	ApiGroup   = "l2crun.tmax.io"
	ApiVersion = "v1"
	L2cKind    = "l2cs"
)

var log = logf.Log.WithName("l2c-apis")
var sonar *sonarqube.SonarQube

func AddV1Apis(parent *wrapper.RouterWrapper, _sonar *sonarqube.SonarQube) error {
	sonar = _sonar
	versionWrapper := wrapper.New(fmt.Sprintf("/%s/%s", ApiGroup, ApiVersion), nil, versionHandler)
	if err := parent.Add(versionWrapper); err != nil {
		return err
	}

	namespaceWrapper := wrapper.New("/namespaces/{namespace}", nil, nil)
	if err := versionWrapper.Add(namespaceWrapper); err != nil {
		return err
	}

	return AddRunApis(namespaceWrapper)
}

func versionHandler(w http.ResponseWriter, _ *http.Request) {
	apiResourceList := &metav1.APIResourceList{}
	apiResourceList.Kind = "APIResourceList"
	apiResourceList.GroupVersion = fmt.Sprintf("%s/%s", ApiGroup, ApiVersion)
	apiResourceList.APIVersion = ApiVersion

	apiResourceList.APIResources = []metav1.APIResource{
		{
			Name:       fmt.Sprintf("%s/run", L2cKind),
			Namespaced: true,
		},
	}

	_ = utils.RespondJSON(w, apiResourceList)
}
