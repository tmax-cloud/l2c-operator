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

	WebhookName     = "global-webhook"
	QualityGateName = "migration"
)

func GetSonarProjectName(l2c *tmaxv1.L2c) string {
	// Project key : <Namespace>_<Name>
	// It's valid as l2c cannot have underscore(_) for its name/namespace
	return fmt.Sprintf("%s_%s", l2c.Namespace, l2c.Name)
}

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

func (s *SonarQube) SetQualityGate() error {
	var gateId int32 = -1
	isGateDefault := false
	// Check if there exists QualityGate
	listResult := &tmaxv1.SonarQubeQualityGateListResult{}
	if err := s.reqHttp(MethodGet, "/api/qualitygates/list", nil, nil, listResult); err != nil {
		return err
	}
	for _, g := range listResult.QualityGates {
		if g.Name == QualityGateName {
			gateId = g.ID
			isGateDefault = g.IsDefault
		}
	}

	// Create QualityGate only if gate is not found
	if gateId < 0 {
		createResult := &tmaxv1.SonarQualityGateCreateResult{}
		if err := s.reqHttp(MethodPost, "/api/qualitygates/create", map[string]string{"name": QualityGateName}, nil, createResult); err != nil {
			return err
		}
		gateId = createResult.ID
		log.Info("Created a QualityGate")
	}

	condMetric := "violations"
	condOP := "GT"
	condError := fmt.Sprintf("%d", 0)

	// Check if condition exists
	showResult := &tmaxv1.SonarQubeQualityGateShowResult{}
	if err := s.reqHttp(MethodGet, "/api/qualitygates/show", map[string]string{"id": fmt.Sprintf("%d", gateId)}, nil, showResult); err != nil {
		return err
	}
	condFound := false
	for _, c := range showResult.Conditions {
		if c.Metric == condMetric && c.OP == condOP && c.Error == condError {
			condFound = true
		}
	}

	// Set condition to the quality gate only if condition is not found
	if !condFound {
		conditionData := map[string]string{
			"gateId": fmt.Sprintf("%d", gateId),
			"metric": condMetric,
			"op":     condOP,
			"error":  condError,
		}
		if err := s.reqHttp(MethodPost, "/api/qualitygates/create_condition", conditionData, nil, nil); err != nil {
			return err
		}
		log.Info("Created a QualityGate Condition")
	}

	// Set as default only if the gate is not a default
	if !isGateDefault {
		if err := s.reqHttp(MethodPost, "/api/qualitygates/set_as_default", map[string]string{"id": fmt.Sprintf("%d", gateId)}, nil, nil); err != nil {
			return err
		}
		log.Info("Set the QualityGate as default")
	}

	return nil
}

func (s *SonarQube) CreateProject(l2c *tmaxv1.L2c) error {
	name := GetSonarProjectName(l2c)
	// Search if there is project
	getResult := &tmaxv1.SonarProjectResult{}
	if err := s.reqHttp(MethodGet, "/api/projects/search", map[string]string{"projects": name}, nil, getResult); err != nil {
		return err
	}

	if len(getResult.Components) > 0 {
		return nil
	}

	// Create project
	data := map[string]string{
		"name":    name,
		"project": name,
	}
	if err := s.reqHttp(MethodPost, "/api/projects/create", data, nil, nil); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Created SonarQube project [%s]", name))

	return nil
}

func (s *SonarQube) DeleteProject(l2c *tmaxv1.L2c) error {
	name := GetSonarProjectName(l2c)
	// Search if there is project
	getResult := &tmaxv1.SonarProjectResult{}
	if err := s.reqHttp(MethodGet, "/api/projects/search", map[string]string{"projects": name}, nil, getResult); err != nil {
		return err
	}

	if len(getResult.Components) == 0 {
		return nil
	}

	if err := s.reqHttp(MethodPost, "/api/projects/delete", map[string]string{"project": name}, nil, nil); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Deleted SonarQube project [%s]", name))

	return nil
}

func (s *SonarQube) GetQualityProfiles(profileNames []string) ([]tmaxv1.SonarProfile, error) {
	profileSelector := ""
	if profileNames != nil {
		profileSelector = strings.Join(profileNames, ",")
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
func (s *SonarQube) SetQualityProfiles(l2c *tmaxv1.L2c, sourceWas string) error {
	name := GetSonarProjectName(l2c)
	// QualityProfile name - temporarily, same as targetWas
	qualityProfile := sourceWas

	// Get QualityProfile List first
	profiles, err := s.GetQualityProfiles([]string{qualityProfile})
	if err != nil {
		return err
	}

	// Get Set QualityProfiles to the project
	listResult := &tmaxv1.SonarQubeQualityProfileListResult{}
	if err := s.reqHttp(MethodGet, "/api/qualityprofiles/search", map[string]string{"project": name}, nil, listResult); err != nil {
		return err
	}

	// Set QualityProfiles for each found getResult
	for _, p := range profiles {
		// Check if profile is already set
		isSet := false
		for _, setP := range listResult.Profiles {
			if p.Language == setP.Language && p.Key == setP.Key {
				isSet = true
				break
			}
		}
		if isSet {
			continue
		}

		data := map[string]string{
			"language":       p.Language,
			"project":        name,
			"qualityProfile": qualityProfile,
		}
		if err := s.reqHttp(MethodPost, "/api/qualityprofiles/add_project", data, nil, nil); err != nil {
			return err
		}

		log.Info(fmt.Sprintf("Set SonarQube Porject %s quality profile %s/%s", name, p.Language, qualityProfile))
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
		if w.Name != WebhookName {
			continue
		}

		// Same name & Same addr -> don't need to do anything
		if w.URL == addr {
			return nil
		}

		// If same name & diff addr, update it
		if err := s.UpdateWebhook(w.Key, addr); err != nil {
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

func (s *SonarQube) GetIssues(key string) ([]tmaxv1.SonarIssue, error) {
	data := map[string]string{
		"componentKeys": key,
		"resolved":      "false",
		"ps":            "500",
	}

	issueList := &tmaxv1.SonarIssueResult{}
	if err := s.reqHttp(MethodGet, "/api/issues/search", data, nil, issueList); err != nil {
		return nil, err
	}

	return issueList.Issues, nil
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
