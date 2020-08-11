package sonarqube

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
)

const (
	MethodPost = "POST"
	MethodGet  = "GET"
	MethodPut  = "PUT"
)

func (s *SonarQube) GenerateToken() (string, error) {
	tokenName := s.AdminId + utils.RandString(5)
	data := map[string]string{
		"login": s.AdminId,
		"name":  tokenName,
	}
	resp, err := s.reqHttp(MethodPost, "/api/user_tokens/generate", data, nil)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	resultBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf(string(resultBytes))
	}

	result := &Token{}
	if err := json.Unmarshal(resultBytes, result); err != nil {
		return "", err
	}

	log.Info(fmt.Sprintf("SonarQube access token %s is generated", tokenName))

	return result.Token, nil
}

func (s *SonarQube) ChangePassword(new string) error {
	data := map[string]string{
		"login":            s.AdminId,
		"previousPassword": s.AdminPw,
		"password":         new,
	}
	resp, err := s.reqHttp(MethodPost, "/api/users/change_password", data, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	resultBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error(err, fmt.Sprintf("hmm %d", resp.StatusCode))
		return fmt.Errorf(string(resultBytes))
	}

	log.Info("SonarQube password is changed")

	return nil
}

func (s *SonarQube) reqHttp(method string, path string, data map[string]string, header map[string]string) (*http.Response, error) {
	uri, err := url.Parse(s.URL + path)
	if err != nil {
		log.Error(err, "PARSE")
		return nil, err
	}

	var bodyReader io.Reader

	// Query or Body
	if data != nil {
		params := url.Values{}
		for k, v := range data {
			params.Set(k, v)
		}
		encoded := params.Encode()

		// Query string if it's get
		if strings.ToLower(method) == "get" {
			uri.RawQuery = encoded
		} else {
			bodyReader = strings.NewReader(params.Encode())
		}
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, uri.String(), bodyReader)
	if err != nil {
		log.Error(err, "REQ")
		return nil, err
	}

	// Header
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}

	// Auth
	if s.Token == "" {
		req.SetBasicAuth(s.AdminId, s.AdminPw)
	} else {
		req.SetBasicAuth(s.Token, "")
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err, "RESP")
		return nil, err
	}

	return resp, nil
}
