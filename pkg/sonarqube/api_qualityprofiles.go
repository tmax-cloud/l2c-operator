package sonarqube

import (
	"fmt"
	"net/http"
	"strings"

	tmaxv1 "github.com/tmax-cloud/l2c-operator/pkg/apis/tmax/v1"
)

const (
	EmptyProfileName = "empty"

	RuleKeyXmlDisallowTomcat   = "xml:TomcatDependXMLCheck"
	RuleKeyXmlDisallowWeblogic = "xml:WeblogicDependXMLCheck"

	RuleKeyJavaDisallowWeblogic = "tmaxsoft-java:DisallowWeblogicDependency"
	RuleKeyJavaDisallowOracle   = "mycompany-java:DBMigrationInvestigation"
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

func (s *SonarQube) CreateQualityProfile(name, lang string) (*tmaxv1.SonarProfile, error) {
	data := map[string]string{
		"name":     name,
		"language": lang,
	}
	result := &tmaxv1.SonarProfileCreateResult{}
	if err := s.reqHttp(http.MethodPost, "/api/qualityprofiles/create", data, nil, result); err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Created quality profile %s/%s (%s)", lang, name, result.Profile.Key))
	return &result.Profile, nil
}

func (s *SonarQube) DeleteQualityProfile(profile, lang string) error {
	data := map[string]string{
		"qualityProfile": profile,
		"language":       lang,
	}
	return s.reqHttp(http.MethodPost, "/api/qualityprofiles/delete", data, nil, nil)
}

func (s *SonarQube) ActivateQualityProfileRule(profileKey, rule string) error {
	data := map[string]string{
		"key":      profileKey,
		"rule":     rule,
		"severity": "BLOCKER",
	}

	return s.reqHttp(http.MethodPost, "/api/qualityprofiles/activate_rule", data, nil, nil)
}

func (s *SonarQube) DeactivateQualityProfileRule(profileKey, rule string) error {
	data := map[string]string{
		"key":  profileKey,
		"rule": rule,
	}

	return s.reqHttp(http.MethodPost, "/api/qualityprofiles/deactivate_rule", data, nil, nil)
}

func (s *SonarQube) SetQualityProfiles(l2c *tmaxv1.L2c) error {
	name := l2c.GetSonarProjectName()

	// Get QualityProfile List first
	profiles, err := s.GetQualityProfiles([]string{name})
	if err != nil {
		return err
	}
	profileMap := map[string]tmaxv1.SonarProfile{} // lang-profile map
	for _, p := range profiles {
		profileMap[p.Language] = p
	}

	// Rules to be applied
	desiredRules, err := s.getDesiredRules(l2c)
	if err != nil {
		return err
	}

	desiredRuleMap := map[string][]string{} // lang-rules map
	for _, r := range desiredRules {
		desiredRuleMap[r.Lang] = append(desiredRuleMap[r.Lang], r.Name)
	}

	// For each lang...
	for lang, desiredRules := range desiredRuleMap {
		// Check if Quality Profile exists and create if not
		profileObj, exists := profileMap[lang]
		profile := &profileObj
		// If not exist, create one
		if !exists {
			if profile, err = s.CreateQualityProfile(name, lang); err != nil {
				return err
			}
		}

		// Check if Quality Profile has desired rules activated
		result := &tmaxv1.SonarRuleResult{}
		curRuleMap := map[string]tmaxv1.SonarRule{}
		if err := s.reqHttp(http.MethodGet, "/api/rules/search", map[string]string{"qprofile": profile.Key, "activation": "true"}, nil, result); err != nil {
			return err
		}
		for _, r := range result.Rules {
			curRuleMap[r.Key] = r
		}

		// If not activated, activate
		for _, desiredRuleKey := range desiredRules {
			_, curRuleActivated := curRuleMap[desiredRuleKey]
			if curRuleActivated {
				delete(curRuleMap, desiredRuleKey)
				continue
			}
			if err := s.ActivateQualityProfileRule(profile.Key, desiredRuleKey); err != nil {
				return err
			}
			log.Info(fmt.Sprintf("Activated rule %s to %s", desiredRuleKey, profile.Name))
		}
		// If non-desired rules are activated, deactivate them
		for curRuleKey := range curRuleMap {
			if err := s.DeactivateQualityProfileRule(profile.Key, curRuleKey); err != nil {
				return err
			}
			log.Info(fmt.Sprintf("Deactivated rule %s from %s", curRuleKey, profile.Name))
		}
	}

	// Set QualityProfiles to the project
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
			"qualityProfile": name,
		}
		if err := s.reqHttp(http.MethodPost, "/api/qualityprofiles/add_project", data, nil, nil); err != nil {
			return err
		}

		log.Info(fmt.Sprintf("Set SonarQube Porject %s quality profile %s/%s", name, p.Language, name))
	}

	return nil
}

func (s *SonarQube) DeleteQualityProfiles(l2c *tmaxv1.L2c) error {
	listResult := &tmaxv1.SonarQubeQualityProfileListResult{}
	if err := s.reqHttp(http.MethodGet, "/api/qualityprofiles/search", map[string]string{"qualityProfile": l2c.GetSonarProjectName()}, nil, listResult); err != nil {
		return err
	}

	for _, p := range listResult.Profiles {
		if err := s.reqHttp(http.MethodPost, "/api/qualityprofiles/delete", map[string]string{"language": p.Language, "qualityProfile": p.Name}, nil, nil); err != nil {
			return err
		}
		log.Info(fmt.Sprintf("Deleted Quality Profile %s/%s", p.Language, p.Name))
	}

	return nil
}

type emptyProfileData struct {
	Exist     bool
	IsDefault bool
}

func (s *SonarQube) CreateEmptyQualityProfiles() error {
	langProfiles := map[string]emptyProfileData{}

	// Get all quality profiles
	qualityProfileList := &tmaxv1.SonarQubeQualityProfileListResult{}
	if err := s.reqHttp(http.MethodGet, "/api/qualityprofiles/search", nil, nil, qualityProfileList); err != nil {
		return err
	}
	for _, profile := range qualityProfileList.Profiles {
		var data emptyProfileData
		tmp, exist := langProfiles[profile.Language]
		if exist {
			data = tmp
		} else {
			data = emptyProfileData{Exist: false, IsDefault: false}
		}
		if profile.Name == EmptyProfileName {
			data.Exist = true
			data.IsDefault = profile.IsDefault
		}
		langProfiles[profile.Language] = data
	}

	// Create/Set default empty profile
	for lang, langProfile := range langProfiles {
		if !langProfile.Exist {
			if _, err := s.CreateQualityProfile(EmptyProfileName, lang); err != nil {
				return err
			}
		}
		if !langProfile.IsDefault {
			if err := s.DefaultQualityProfile(EmptyProfileName, lang); err != nil {
				return err
			}
			log.Info(fmt.Sprintf("Set quality profile %s/%s as default", lang, EmptyProfileName))
		}
	}

	return nil
}

func (s *SonarQube) DefaultQualityProfile(name, lang string) error {
	data := map[string]string{
		"qualityProfile": name,
		"language":       lang,
	}
	return s.reqHttp(http.MethodPost, "/api/qualityprofiles/set_default", data, nil, nil)
}

type sonarRule struct {
	Name string
	Lang string
}

func (s *SonarQube) getDesiredRules(l2c *tmaxv1.L2c) ([]sonarRule, error) {
	var rules []sonarRule

	// WAS
	was := l2c.Spec.Was
	// Weblogic -> Jeus
	if was.From.Type == tmaxv1.WasTypeWeblogic && was.To.Type == tmaxv1.WasTypeJeus {
		rules = append(rules, sonarRule{
			Name: RuleKeyJavaDisallowWeblogic,
			Lang: "java",
		}, sonarRule{
			Name: RuleKeyXmlDisallowTomcat,
			Lang: "xml",
		}, sonarRule{
			Name: RuleKeyXmlDisallowWeblogic,
			Lang: "xml",
		})
	}

	// DB
	if l2c.Spec.Db != nil {
		db := l2c.Spec.Db
		// Oracle -> Tibero
		if db.From.Type == tmaxv1.DbTypeOracle && db.To.Type == tmaxv1.DbTypeTibero {
			rules = append(rules, sonarRule{
				Name: RuleKeyJavaDisallowOracle,
				Lang: "java",
			})
		}
	}

	return rules, nil
}
