package sonarqube

import (
	"fmt"
	"net/http"
	"strings"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

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
	if err := s.reqHttp(http.MethodGet, "/api/qualityprofiles/search", data, nil, result); err != nil {
		return nil, err
	}

	return result.Profiles, nil
}

// TODO : qualityProfile name should be revisited - sourceWAS is not enough!
func (s *SonarQube) SetQualityProfiles(l2c *tmaxv1.L2c, sourceWas string) error {
	name := l2c.GetSonarProjectName()
	// QualityProfile name - temporarily, same as targetWas
	qualityProfile := sourceWas

	// Get QualityProfile List first
	profiles, err := s.GetQualityProfiles([]string{qualityProfile})
	if err != nil {
		return err
	}

	// Get Set QualityProfiles to the project
	listResult := &tmaxv1.SonarQubeQualityProfileListResult{}
	if err := s.reqHttp(http.MethodGet, "/api/qualityprofiles/search", map[string]string{"project": name}, nil, listResult); err != nil {
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
		if err := s.reqHttp(http.MethodPost, "/api/qualityprofiles/add_project", data, nil, nil); err != nil {
			return err
		}

		log.Info(fmt.Sprintf("Set SonarQube Porject %s quality profile %s/%s", name, p.Language, qualityProfile))
	}

	return nil
}

type EmptyProfileData struct {
	Exist     bool
	IsDefault bool
}

func (s *SonarQube) CreateEmptyQualityProfiles() error {
	langProfiles := map[string]EmptyProfileData{}

	// Get all quality profiles
	qualityProfileList := &tmaxv1.SonarQubeQualityProfileListResult{}
	if err := s.reqHttp(http.MethodGet, "/api/qualityprofiles/search", nil, nil, qualityProfileList); err != nil {
		return err
	}
	for _, profile := range qualityProfileList.Profiles {
		var data EmptyProfileData
		tmp, exist := langProfiles[profile.Language]
		if exist {
			data = tmp
		} else {
			data = EmptyProfileData{Exist: false, IsDefault: false}
		}
		if profile.Name == "empty" {
			data.Exist = true
			data.IsDefault = profile.IsDefault
		}
		langProfiles[profile.Language] = data
	}

	// Create/Set default empty profile
	for lang, langProfile := range langProfiles {
		if !langProfile.Exist {
			if err := s.CreateEmptyQualityProfile("empty", lang); err != nil {
				return err
			}
			log.Info(fmt.Sprintf("Created quality profile %s/%s", lang, "empty"))
		}
		if !langProfile.IsDefault {
			if err := s.DefaultQualityProfile("empty", lang); err != nil {
				return err
			}
			log.Info(fmt.Sprintf("Set quality profile %s/%s as default", lang, "empty"))
		}
	}

	return nil
}

func (s *SonarQube) CreateEmptyQualityProfile(name, lang string) error {
	data := map[string]string{
		"name":     name,
		"language": lang,
	}
	return s.reqHttp(http.MethodPost, "/api/qualityprofiles/create", data, nil, nil)
}

func (s *SonarQube) DefaultQualityProfile(name, lang string) error {
	data := map[string]string{
		"qualityProfile": name,
		"language":       lang,
	}
	return s.reqHttp(http.MethodPost, "/api/qualityprofiles/set_default", data, nil, nil)
}
