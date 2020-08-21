package sonarqube

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

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

	// Auth
	if s.Token == "" {
		req.SetBasicAuth(s.AdminId, s.AdminPw)
	} else {
		req.SetBasicAuth(s.Token, "")
	}

	// Header
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range header {
		req.Header.Set(k, v)
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
