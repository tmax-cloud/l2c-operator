package v1

//QualityGates
// +k8s:deepcopy-gen=false
type SonarQualityGateCreateResult struct {
	ID int32 `json:"id"`

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarQubeQualityGateListResult struct {
	QualityGates []SonarQubeQualityGate `json:"qualitygates"`

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarQubeQualityGate struct {
	ID        int32  `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarQubeQualityGateShowResult struct {
	Conditions []SonarQubeQualityGateCondition

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarQubeQualityGateCondition struct {
	ID     int32  `json:"id"`
	Metric string `json:"metric"`
	OP     string `json:"op"`
	Error  string `json:"error"`

	X map[string]interface{} `json:"-"`
}

// QualityProfiles
// +k8s:deepcopy-gen=false
type SonarQubeQualityProfileListResult struct {
	Profiles []SonarQubeQualityProfile `json:"profiles"`

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarQubeQualityProfile struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	IsDefault bool   `json:"isDefault"`

	X map[string]interface{} `json:"-"`
}

// Issues
// +k8s:deepcopy-gen=false
type SonarIssueResult struct {
	Issues []SonarIssue `json:"issues"`

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarIssue struct {
	Component    string           `json:"component"`
	Project      string           `json:"project"`
	Line         int32            `json:"line"`
	TextRange    map[string]int32 `json:"textRange"`
	Status       string           `json:"status"`
	Message      string           `json:"message"`
	CreationDate string           `json:"creationDate"`
	UpdateDate   string           `json:"updateDate"`
	Type         string           `json:"type"`
	Severity     string           `json:"severity"`

	X map[string]interface{} `json:"-"`
}

// Users
// +k8s:deepcopy-gen=false
type SonarUserSearchResult struct {
	Users []SonarUser `json:"users"`

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarUser struct {
	Login string `json:"login"`

	X map[string]interface{} `json:"-"`
}

// Groups
// +k8s:deepcopy-gen=false
type SonarGroupSearchResult struct {
	Groups []SonarGroup `json:"groups"`

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarGroup struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`

	X map[string]interface{} `json:"-"`
}

// Tokens
// +k8s:deepcopy-gen=false
type SonarToken struct {
	Token string `json:"token"`

	X map[string]interface{} `json:"-"`
}

// Profiles
// +k8s:deepcopy-gen=false
type SonarProfileResult struct {
	Profiles []SonarProfile `json:"profiles"`

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarProfile struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Language string `json:"language"`

	X map[string]interface{} `json:"-"`
}

// Projects
// +k8s:deepcopy-gen=false
type SonarProjectResult struct {
	Components []interface{} `json:"components"`

	X map[string]interface{} `json:"-"`
}

// Webhooks
// +k8s:deepcopy-gen=false
type SonarWebhookResult struct {
	Webhooks []SonarWebhook `json:"webhooks"`

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarWebhook struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarWebhookRequest struct {
	Project     SonarWebhookProject     `json:"project"`
	QualityGate SonarWebhookQualityGate `json:"qualityGate"`

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarWebhookProject struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`

	X map[string]interface{} `json:"-"`
}

// +k8s:deepcopy-gen=false
type SonarWebhookQualityGate struct {
	Name   string `json:"name"`
	Status string `json:"status"`

	X map[string]interface{} `json:"-"`
}
