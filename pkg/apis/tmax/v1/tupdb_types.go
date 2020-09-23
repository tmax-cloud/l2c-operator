package v1

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TupDBSpec defines the desired state of TupDB
type TupDBSpec struct {
	// DataBase Information

	// DB Source configuration
	From TupDBFrom `json:"from"`

	// DB destination configuration
	To TupDBTo `json:"to"`
}

type TupDBFrom struct {
	// Current DB Type
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

type TupDBTo struct {
	// Target DB type, to be migrated
	// +kubebuilder:validation:Enum=tibero
	Type string `json:"type,omitempty"`

	// Storage size of target DB
	StorageSize string `json:"storageSize,omitempty"`

	// User for target DB
	User string `json:"user,omitempty"`

	// Password for target DB
	Password string `json:"password,omitempty"`

	// Current DB SID
	Sid string `json:"sid,omitempty"`
}

// TupDBStatus defines the observed state of TupDB
type TupDBStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Conditions []status.Condition `json:"conditions,omitempty"`

	LastAnalyzeResult string `json:"lastAnalyzeResult,omitempty"`

	// Target DB host
	// +kubebuilder:validation:Pattern=(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])
	TargetHost string `json:"targetHost,omitempty"`

	// Target DB port
	TargetPort int32 `json:"targetPort,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TupDB is the Schema for the tupdbs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=tupdbs,scope=Namespaced
type TupDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TupDBSpec   `json:"spec,omitempty"`
	Status TupDBStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TupDBList contains a list of TupDB
type TupDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TupDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TupDB{}, &TupDBList{})
}
