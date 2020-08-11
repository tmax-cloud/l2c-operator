package v1

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

type SonarToken struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Token     string `json:"token"`
	CreatedAt string `json:"createdAt"`
}