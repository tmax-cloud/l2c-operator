package v1

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TupWASSpec defines the desired state of TupWAS
type TupWASSpec struct {
	// WAS source configuration
	From TupWasFrom `json:"from"`

	// WAS destination configuration
	To TupWasTo `json:"to"`
}

type TupWasGit struct {
	// URL of git repository
	Url string `json:"url"`

	// Revision to be used as a source
	Revision string `json:"revision,omitempty"`
}

type TupWasImage struct {
	// Image URL where the built application image is stored
	Url string `json:"url"`

	// Secret name that contains a credential to access registry, if the image registry needs credentials to push or pull an image
	RegSecret string `json:"regSecret,omitempty"`
}

type TupWasFrom struct {
	// Current WAS type
	// +kubebuilder:validation:Enum=weblogic
	Type string `json:"type"`

	// Git information for WAS source code
	Git TupWasGit `json:"git"`

	// Package server URL that would be used while building the application
	PackageServer string `json:"packageServerUrl,omitempty"`
}

type TupWasTo struct {
	// Target WAS type, to be migrated
	// +kubebuilder:validation:Enum="jeus:7";"jeus:8"
	Type string `json:"type"`

	// Image, in which the built application image would be saved
	Image TupWasImage `json:"image"`

	// Port number WAS would use
	Port int32 `json:"port"`
}

// TupWASStatus defines the observed state of TupWAS
type TupWASStatus struct {
	// Start time of last analysis
	LastAnalyzeStartTime *metav1.Time `json:"lastAnalyzeStartTime,omitempty"`

	// Completion time of last analysis
	LastAnalyzeCompletionTime *metav1.Time `json:"lastAnalyzeCompletionTime,omitempty"`

	// Result of last analysis
	LastAnalyzeResult string `json:"lastAnalyzeResult,omitempty"`

	// Start time of last build
	LastBuildStartTime *metav1.Time `json:"lastBuildStartTime,omitempty"`

	// Completion time of last build
	LastBuildCompletionTime *metav1.Time `json:"lastBuildCompletionTime,omitempty"`

	// Result of last build
	LastBuildResult string `json:"lastBuildResult,omitempty"`

	// PipelineRun name for Analyze
	AnalyzePipelineRunName string `json:"analyzePipelineRunName,omitempty"`

	// PipelineRun name for Build/Deploy
	BuildPipelineRunName string `json:"buildPipelineRunName,omitempty"`

	// TupWAS project conditions
	Conditions []status.Condition `json:"conditions,omitempty"`

	// Editor (VSCode) status
	Editor *EditorStatus `json:"editor,omitempty"`

	// T-up Jeus URL
	ReportUrl string `json:"reportUrl,omitempty"`

	// Migrated Was URL
	WasUrl string `json:"wasUrl,omitempty"`
}

type EditorStatus struct {
	// VSCode URL
	Url string `json:"url,omitempty"`

	// VSCode access code
	Password string `json:"password,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TupWAS is the Schema for the tupwas API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=tupwas,scope=Namespaced
type TupWAS struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TupWASSpec   `json:"spec,omitempty"`
	Status TupWASStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TupWASList contains a list of TupWAS
type TupWASList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TupWAS `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TupWAS{}, &TupWASList{})
}
