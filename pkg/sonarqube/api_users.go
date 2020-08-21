package sonarqube

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/tmax-cloud/l2c-operator/internal/utils"
	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

func (s *SonarQube) GenerateToken(id, pw string) (string, error) {
	tokenName := id + utils.RandString(5)
	data := map[string]string{
		"login": id,
		"name":  tokenName,
	}
	header := map[string]string{
		"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", id, pw))),
	}

	result := &tmaxv1.SonarToken{}
	if err := s.reqHttp(http.MethodPost, "/api/user_tokens/generate", data, header, result); err != nil {
		return "", err
	}

	log.Info(fmt.Sprintf("SonarQube access token %s is generated", tokenName))

	return result.Token, nil
}

func (s *SonarQube) ChangeAdminPassword(new string) error {
	data := map[string]string{
		"login":            s.AdminId,
		"previousPassword": s.AdminPw,
		"password":         new,
	}
	if err := s.reqHttp(http.MethodPost, "/api/users/change_password", data, nil, nil); err != nil {
		return err
	}

	log.Info("SonarQube password is changed")

	return nil
}

func (s *SonarQube) GetGroup(group string) (*tmaxv1.SonarGroup, error) {
	groupResult := &tmaxv1.SonarGroupSearchResult{}
	if err := s.reqHttp(http.MethodGet, "/api/permissions/groups", nil, nil, groupResult); err != nil {
		return nil, err
	}

	for _, g := range groupResult.Groups {
		if g.Name == group {
			return &g, nil
		}
	}

	return nil, fmt.Errorf("cannot find group %s", group)
}

func (s *SonarQube) RemoveAllGroupPermissions(group string) error {
	groupInfo, err := s.GetGroup(group)
	if err != nil {
		return err
	}
	if len(groupInfo.Permissions) == 0 {
		return nil
	}

	permissions := []string{"admin", "profileadmin", "gateadmin", "scan", "provisioning"}
	data := map[string]string{
		"groupName": group,
	}
	for _, p := range permissions {
		data["permission"] = p
		if err := s.reqHttp(http.MethodPost, "/api/permissions/remove_group", data, nil, nil); err != nil {
			return err
		}
		log.Info(fmt.Sprintf("Removed group permission %s for group %s", p, group))
	}
	return nil
}

func (s *SonarQube) CreateUser(name, pw string) error {
	userResult := &tmaxv1.SonarUserSearchResult{}
	if err := s.reqHttp(http.MethodGet, "/api/users/search", map[string]string{"q": name}, nil, userResult); err != nil {
		return err
	}

	if len(userResult.Users) > 0 {
		return nil
	}

	data := map[string]string{
		"login":    name,
		"name":     name,
		"password": pw,
	}

	if err := s.reqHttp(http.MethodPost, "/api/users/create", data, nil, nil); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Created user %s", name))

	return nil
}

func (s *SonarQube) AddUserPermission(name, permission string) error {
	data := map[string]string{
		"login":      name,
		"permission": permission,
	}

	if err := s.reqHttp(http.MethodPost, "/api/permissions/add_user", data, nil, nil); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Added permission %s to user %s", permission, name))

	return nil
}
