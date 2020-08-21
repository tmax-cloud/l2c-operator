package sonarqube

import (
	"fmt"
	"net/http"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

func (s *SonarQube) CreateProject(l2c *tmaxv1.L2c) error {
	name := l2c.GetSonarProjectName()
	// Search if there is project
	getResult := &tmaxv1.SonarProjectResult{}
	if err := s.reqHttp(http.MethodGet, "/api/projects/search", map[string]string{"projects": name}, nil, getResult); err != nil {
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
	if err := s.reqHttp(http.MethodPost, "/api/projects/create", data, nil, nil); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Created SonarQube project [%s]", name))

	return nil
}

func (s *SonarQube) DeleteProject(l2c *tmaxv1.L2c) error {
	name := l2c.GetSonarProjectName()
	// Search if there is project
	getResult := &tmaxv1.SonarProjectResult{}
	if err := s.reqHttp(http.MethodGet, "/api/projects/search", map[string]string{"projects": name}, nil, getResult); err != nil {
		return err
	}

	if len(getResult.Components) == 0 {
		return nil
	}

	if err := s.reqHttp(http.MethodPost, "/api/projects/delete", map[string]string{"project": name}, nil, nil); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Deleted SonarQube project [%s]", name))

	return nil
}
