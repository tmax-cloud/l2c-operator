package v1

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// L2cSpec defines the desired state of L2c
type L2cSpec struct {
	// Was migration configuration
	Was *L2cWas `json:"was"`

	// Db migration configuration
	Db *L2cDb `json:"db,omitempty"`
}

type L2cWas struct {
	// Package server URL that would be used while building the application
	WasPackageServer string `json:"wasPackageServerUrl,omitempty"`

	// WAS source configuration
	From *L2cWasFrom `json:"from"`

	// WAS destination configuration
	To *L2cWasTo `json:"to"`
}

type L2cGit struct {
	// URL of git repository
	Url string `json:"url"`

	// Revision to be used as a source
	Revision string `json:"revision,omitempty"`
}

type L2cImage struct {
	// Image URL where the built application image is stored
	Url string `json:"url"`

	// Secret name that contains a credential to access registry, if the image registry needs credentials to push or pull an image
	RegSecret string `json:"regSecret,omitempty"`
}

type L2cWasFrom struct {
	// Current WAS type
	// +kubebuilder:validation:Enum=wildfly
	Type string `json:"type"`

	// Git information for WAS source code
	Git *L2cGit `json:"git"`
}

type L2cWasTo struct {
	// Target WAS type, to be migrated
	// +kubebuilder:validation:Enum=jeus
	Type string `json:"type"`

	// Image, in which the built application image would be saved
	Image *L2cImage `json:"image"`

	// Port number WAS would use
	Port int32 `json:"port"`

	// Service type WAS would use
	// +kubebuilder:validation:Enum=ClusterIP;LoadBalancer;NodePort
	ServiceType string `json:"serviceType,omitempty"`
}

type L2cDb struct {
	// DB source configuration
	From *L2cDbFrom `json:"from"`

	// DB destination configuration
	To *L2cDbTo `json:"to"`
}

type L2cDbFrom struct {
	// Current DB type
	// +kubebuilder:validation:Enum=oracle
	Type string `json:"type,omitempty"`

	// Current DB host
	// +kubebuilder:validation:Pattern=(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])
	Host string `json:"host,omitempty"`

	// Current DB port
	Port int32 `json:"port,omitempty"`

	// Current DB user
	User string `json:"user,omitempty"`

	// Current DB password
	Password string `json:"password,omitempty"`

	// Current DB SID
	Sid string `json:"sid,omitempty"`
}

type L2cDbTo struct {
	// Target DB type, to be migrated
	// +kubebuilder:validation:Enum=tibero
	Type string `json:"type,omitempty"`

	// Storage size of target DB
	StorageSize string `json:"storageSize,omitempty"`

	// User for target DB
	User string `json:"user,omitempty"`

	// Password for target DB
	Password string `json:"password,omitempty"`
}

// L2cStatus defines the observed state of L2c
type L2cStatus struct {
	// Completed timestamp
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Pipeline name for the L2c
	PipelineName string `json:"pipelineName,omitempty"`

	// PipelineRun name for the L2c
	PipelineRunName string `json:"pipelineRunName,omitempty"`

	Conditions []status.Condition `json:"conditions"`

	// Status of each Task
	TaskStatus []L2cTaskStatus `json:"taskStatus"`

	// VSCode URL
	EditorUrl string `json:"editorUrl,omitempty"`

	// VSCode access code
	EditorCode string `json:"editorCode,omitempty"`

	// SonarQube issues
	SonarIssues []CodeIssue `json:"sonarIssues"`
}

type CodeIssue struct {
}

type L2cTaskStatus struct {
	//
	TaskRunName string `json:"taskRunName"`

	//
	Conditions []status.Condition `json:"conditions"`

	//
	PodName string `json:"podName"`

	//
	StartTime *metav1.Time `json:"startTime,omitempty"`

	//
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	//
	Steps []tektonv1.StepState `json:"steps,omitempty"`

	//
	Sidecars []tektonv1.SidecarState `json:"sidecars,omitempty"`

	// TaskSpec contains the Spec from the dereferenced Task definition used to instantiate this TaskRun.
	TaskSpec *tektonv1.TaskSpec `json:"taskSpec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// L2c is the Schema for the l2cs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=l2cs,scope=Namespaced
type L2c struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   L2cSpec   `json:"spec,omitempty"`
	Status L2cStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// L2cList contains a list of L2c
type L2cList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []L2c `json:"items"`
}

func init() {
	SchemeBuilder.Register(&L2c{}, &L2cList{})
}
