package v1

// Issues
type SonarIssueResult struct {
	Total       int32               `json:"total"`
	P           int32               `json:"p"`
	Ps          int32               `json:"ps"`
	Paging      map[string]int32    `json:"paging"`
	EffortTotal int32               `json:"effortTotal"`
	DebtTotal   int32               `json:"deptTotal"`
	Issues      []SonarIssue        `json:"issues"`
	Components  []map[string]string `json:"components"`
}

type SonarIssue struct {
	Key          string              `json:"key"`
	Component    string              `json:"component"`
	Project      string              `json:"project"`
	Organization string              `json:"organization"`
	Rule         string              `json:"rule"`
	Status       string              `json:"status"`
	Resolution   string              `json:"resolution"`
	Severity     string              `json:"severity"`
	Message      string              `json:"message"`
	Line         int32               `json:"line"`
	Hash         string              `json:"hash"`
	Author       string              `json:"author"`
	Effort       string              `json:"effort"`
	Dept         string              `json:"dept"`
	CreationDate string              `json:"creationDate"`
	UpdateDate   string              `json:"updateDate"`
	Tags         []string            `json:"tags"`
	Type         string              `json:"type"`
	Comments     []map[string]string `json:"comments"`
	Attr         map[string]string   `json:"attr"`
	Transitions  []string            `json:"transitions"`
	Actions      []string            `json:"actions"`
	TextRange    map[string]int32    `json:"textRange"`
	Flows        []SonarFlow         `json:"flows"`
}

type SonarFlow struct {
	Locations []SonarLocation `json:"locations"`
}

type SonarLocation struct {
	Message   string           `json:"msg"`
	TextRange map[string]int32 `json:"textRange"`
}

// Tokens
type SonarToken struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Token     string `json:"token"`
	CreatedAt string `json:"createdAt"`
}

// Profiles
type SonarProfileResult struct {
	Profiles []SonarProfile  `json:"profiles"`
	Actions  map[string]bool `json:"actions"`
}

type SonarProfile struct {
	Key                       string          `json:"key"`
	Name                      string          `json:"name"`
	Language                  string          `json:"language"`
	LanguageName              string          `json:"languageName"`
	IsInherited               bool            `json:"isInherited"`
	IsDefault                 bool            `json:"isDefault"`
	ActiveRuleCount           int32           `json:"activeRuleCount"`
	ActiveDeprecatedRuleCount int32           `json:"activeDeprecatedRuleCount"`
	RulesUpdatedAt            string          `json:"rulesUpdatedAt"`
	Organization              string          `json:"organization"`
	IsBuiltIn                 bool            `json:"isBuiltIn"`
	Actions                   map[string]bool `json:"actions"`
}

// Projects
type SonarProjectResult struct {
	Paging     map[string]int32 `json:"paging"`
	Components []SonarProject   `json:"components"`
}

type SonarProject struct {
	Organization     string `json:"organization"`
	Key              string `json:"key"`
	Name             string `json:"name"`
	Qualifier        string `json:"qualifier"`
	Visibility       string `json:"visibility"`
	LastAnalysisDate string `json:"lastAnalysisDate"`
	Revision         string `json:"revision"`
}

// Webhooks
type SonarWebhookResult struct {
	Webhooks []map[string]string `json:"webhooks"`
}
