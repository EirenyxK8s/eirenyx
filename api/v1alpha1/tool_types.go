package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

type ToolType string

const (
	ToolTrivy  ToolType = "trivy"
	ToolFalco  ToolType = "falco"
	ToolLitmus ToolType = "litmus"
)

const ToolFinalizer = "eirenyx.tool/finalizer"

// ToolSpec defines the desired state of Tool
type ToolSpec struct {
	Type      ToolType `json:"type"`
	Enabled   bool     `json:"enabled"`
	Namespace string   `json:"namespace,omitempty"`

	// Values contain Helm values passed directly to the tool chart.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Values runtime.RawExtension `json:"values,omitempty"`
}

// ToolStatus defines the observed state of the Tool.
type ToolStatus struct {
	Installed bool   `json:"installed,omitempty"`
	Healthy   bool   `json:"healthy,omitempty"`
	Version   string `json:"version,omitempty"`

	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Tool is the Schema for the tools API
type Tool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitzero"`

	Spec ToolSpec `json:"spec"`
	// +optional
	Status ToolStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// ToolList contains a list of Tool
type ToolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`

	Items []Tool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tool{}, &ToolList{})
}
