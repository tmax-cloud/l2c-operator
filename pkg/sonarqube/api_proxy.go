package sonarqube

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
)

// Read-only proxy for SonarQube server
func (s *SonarServer) ApiProxyHandleFunc(w http.ResponseWriter, r *http.Request) {
	// If method is not GET, reject!
	if r.Method != http.MethodGet {
		_ = utils.RespondError(w, http.StatusBadRequest, "it's be read-only, only GET is permitted")
		return
	}

	// TODO : should we do some access control for accessing project / setting or something??

	// Read from request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer r.Body.Close()
	bodyReader := bytes.NewReader(body)

	// Forward to actual SonarQube server with token
	uri, err := url.Parse(s.sonar.URL + r.URL.String())
	if err != nil {
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(r.Method, uri.String(), bodyReader)
	if err != nil {
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	req.SetBasicAuth(s.sonar.Token, "")

	// Read from response
	resp, err := client.Do(req)
	if err != nil {
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()
	resultBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Forward response to end user
	_, err = w.Write(resultBytes)
	if err != nil {
		_ = utils.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
}
