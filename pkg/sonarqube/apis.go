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
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

const (
	MethodPost = "POST"
	MethodGet  = "GET"

	WebhookName = "global-webhook"
)

func (s *SonarQube) GenerateToken() (string, error) {
	tokenName := s.AdminId + utils.RandString(5)
	data := map[string]string{
		"login": s.AdminId,
		"name":  tokenName,
	}
	result := &tmaxv1.SonarToken{}
	if err := s.reqHttp(MethodPost, "/api/user_tokens/generate", data, nil, result); err != nil {
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
	if err := s.reqHttp(MethodPost, "/api/users/change_password", data, nil, nil); err != nil {
		return err
	}

	log.Info("SonarQube password is changed")

	return nil
}

func (s *SonarQube) CreateProject(name string) error {
	key := utils.ToAlphaNumeric(name)

	// Search if there is project
	getResult := &tmaxv1.SonarProjectResult{}
	if err := s.reqHttp(MethodGet, "/api/projects/search", map[string]string{"projects": key}, nil, getResult); err != nil {
		return err
	}

	if len(getResult.Components) > 0 {
		return nil
	}

	// Create project
	data := map[string]string{
		"name":    name,
		"project": key,
	}
	if err := s.reqHttp(MethodPost, "/api/projects/create", data, nil, nil); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Created SonarQube project [%s/%s]", name, key))

	return nil
}

func (s *SonarQube) DeleteProject(name string) error {
	key := utils.ToAlphaNumeric(name)

	if err := s.reqHttp(MethodPost, "/api/projects/delete", map[string]string{"project": key}, nil, nil); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Deleted SonarQube project [%s/%s]", name, key))

	return nil
}

func (s *SonarQube) GetQualityProfiles(names []string) ([]tmaxv1.SonarProfile, error) {
	profileSelector := ""
	if names != nil {
		profileSelector = strings.Join(names, ",")
	}

	data := map[string]string{}
	if profileSelector != "" {
		data["qualityProfile"] = profileSelector
	}

	result := &tmaxv1.SonarProfileResult{}
	if err := s.reqHttp(MethodGet, "/api/qualityprofiles/search", data, nil, result); err != nil {
		return nil, err
	}

	return result.Profiles, nil
}

// TODO : qualityProfile name should be revisited - sourceWAS is not enough!
func (s *SonarQube) SetQualityProfiles(projectName string, sourceWas string) error {
	key := utils.ToAlphaNumeric(projectName)

	// QualityProfile name - temporarily, same as targetWas
	qualityProfile := sourceWas

	// Get QualityProfile List first
	profiles, err := s.GetQualityProfiles([]string{qualityProfile})
	if err != nil {
		return err
	}

	// Set QualityProfiles for each found getResult
	for _, p := range profiles {
		data := map[string]string{
			"language":       p.Language,
			"project":        key,
			"qualityProfile": qualityProfile,
		}
		if err := s.reqHttp(MethodPost, "/api/qualityprofiles/add_project", data, nil, nil); err != nil {
			return err
		}

		log.Info(fmt.Sprintf("Set SonarQube Porject %s quality profile %s/%s", key, p.Language, qualityProfile))
	}

	return nil
}

func (s *SonarQube) RegisterWebhook() error {
	addr := fmt.Sprintf("http://l2c-operator:%d", Port)

	// First, get if webhook is already set correctly
	getResult := &tmaxv1.SonarWebhookResult{}
	if err := s.reqHttp(MethodPost, "/api/webhooks/list", nil, nil, getResult); err != nil {
		return err
	}

	for _, w := range getResult.Webhooks {
		name, exist := w["name"]
		if !exist {
			return fmt.Errorf("invalid webhook return %+v", w)
		}
		if name != WebhookName {
			continue
		}

		uri, exist := w["url"]
		if !exist {
			return fmt.Errorf("invalid webhook return %+v", w)
		}
		// Same name & Same addr -> don't need to do anything
		if uri == addr {
			return nil
		}

		// If same name & diff addr, update it
		key, exist := w["key"]
		if !exist {
			return fmt.Errorf("invalid webhook return %+v", w)
		}
		if err := s.UpdateWebhook(key, addr); err != nil {
			return err
		}
		return nil
	}

	// Register webhook
	data := map[string]string{
		"name": WebhookName,
		"url":  addr,
	}

	if err := s.reqHttp(MethodPost, "/api/webhooks/create", data, nil, nil); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Set SonarQube Webhook global-webhook as %s", addr))

	return nil
}

func (s *SonarQube) UpdateWebhook(key, uri string) error {
	data := map[string]string{
		"name":    WebhookName,
		"webhook": key,
		"url":     uri,
	}
	if err := s.reqHttp(MethodPost, "/api/webhooks/update", data, nil, nil); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Updates SonarQube Webhook global-webhook as %s", uri))

	return nil
}

func (s *SonarQube) reqHttp(method string, path string, data map[string]string, header map[string]string, handledResp interface{}) error {
	uri, err := url.Parse(s.URL + path)
	if err != nil {
		return err
	}

	// Query or Body
	params := url.Values{}
	for k, v := range data {
		params.Set(k, v)
	}
	encoded := params.Encode()

	// Query string if it's get, else, body
	var bodyReader io.Reader
	if strings.ToLower(method) == "get" {
		uri.RawQuery = encoded
	} else {
		bodyReader = strings.NewReader(params.Encode())
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, uri.String(), bodyReader)
	if err != nil {
		return err
	}

	// Header
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}

	// Auth
	if s.Token == "" {
		req.SetBasicAuth(s.AdminId, s.AdminPw)
	} else {
		req.SetBasicAuth(s.Token, "")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if err := handleResp(resp, handledResp); err != nil {
		return err
	}

	return nil
}

func handleResp(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()
	resultBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error(fmt.Errorf("status code:  %d", resp.StatusCode), "")
		return fmt.Errorf(string(resultBytes))
	}

	if result == nil {
		return nil
	}

	if err := json.Unmarshal(resultBytes, result); err != nil {
		return err
	}

	return nil
}
