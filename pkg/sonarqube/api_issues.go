package sonarqube

import (
	"net/http"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

func (s *SonarQube) GetIssues(key string) ([]tmaxv1.SonarIssue, error) {
	data := map[string]string{
		"componentKeys": key,
		"resolved":      "false",
		"ps":            "500",
	}

	issueList := &tmaxv1.SonarIssueResult{}
	if err := s.reqHttp(http.MethodGet, "/api/issues/search", data, nil, issueList); err != nil {
		return nil, err
	}

	return issueList.Issues, nil
}
