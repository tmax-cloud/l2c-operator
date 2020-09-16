package v1

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	authorization "k8s.io/api/authorization/v1"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
)

const (
	UserHeader   = "X-Remote-User"
	GroupHeader  = "X-Remote-Group"
	ExtrasHeader = "X-Remote-Extra-"
)

func Authorize(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := authorize(w, req); err != nil {
			_ = utils.RespondError(w, http.StatusUnauthorized, err.Error())
			return
		}

		if err := reviewAccess(w, req); err != nil {
			_ = utils.RespondError(w, http.StatusUnauthorized, err.Error())
			return
		}

		h.ServeHTTP(w, req)
	})
}

func authorize(w http.ResponseWriter, req *http.Request) error {
	if req.TLS == nil || len(req.TLS.PeerCertificates) == 0 {
		_ = utils.RespondError(w, http.StatusBadRequest, "is not https or there is no peer certificate")
		return fmt.Errorf("")
	}
	return nil
}

func reviewAccess(w http.ResponseWriter, req *http.Request) error {
	userName, err := getUserName(req.Header)
	if err != nil {
		return err
	}

	userGroups, err := getUserGroup(req.Header)
	if err != nil {
		return err
	}

	userExtras := getUserExtras(req.Header)

	// URL : /apis/tup.tmax.io/v1/namespaces/<namespace>/[tupwas|tupdbs]/<resource name>/[analyze|run]
	subPaths := strings.Split(req.URL.Path, "/")
	if len(subPaths) != 9 {
		return fmt.Errorf("URL should be in form of '/apis/tup.tmax.io/v1/namespaces/<namespace>/[tupwas|tupdbs]/<resource name>/[analyze|run]'")
	}
	resource := subPaths[6]
	subResource := subPaths[8]

	vars := mux.Vars(req)

	ns, nsExist := vars["namespace"]
	resourceName, nameExist := vars["tupName"]
	if !nsExist || !nameExist {
		_ = utils.RespondError(w, http.StatusBadRequest, "url is malformed")
		return fmt.Errorf("")
	}

	r := &authorization.SubjectAccessReview{
		Spec: authorization.SubjectAccessReviewSpec{
			User:   userName,
			Groups: userGroups,
			Extra:  userExtras,
			ResourceAttributes: &authorization.ResourceAttributes{
				Name:        resourceName,
				Namespace:   ns,
				Group:       ApiGroup,
				Version:     ApiVersion,
				Resource:    resource,
				Subresource: subResource,
				Verb:        "update",
			},
		},
	}

	authCli, err := utils.AuthClient()
	if err != nil {
		return err
	}

	result, err := authCli.SubjectAccessReviews().Create(r)
	if err != nil {
		return err
	}

	if result.Status.Allowed {
		return nil
	}

	return fmt.Errorf(result.Status.Reason)
}

func getUserName(header http.Header) (string, error) {
	for k, v := range header {
		if k == UserHeader {
			return v[0], nil
		}
	}
	return "", fmt.Errorf("no header %s", UserHeader)
}

func getUserGroup(header http.Header) ([]string, error) {
	for k, v := range header {
		if k == UserHeader {
			return v, nil
		}
	}
	return nil, fmt.Errorf("no header %s", GroupHeader)
}

func getUserExtras(header http.Header) map[string]authorization.ExtraValue {
	extras := map[string]authorization.ExtraValue{}

	for k, v := range header {
		if strings.HasPrefix(k, ExtrasHeader) {
			extras[strings.TrimPrefix(k, ExtrasHeader)] = v
		}
	}

	return extras
}
