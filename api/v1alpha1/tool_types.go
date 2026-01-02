package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ToolType string

const (
	ToolTrivy  ToolType = "trivy"
	ToolFalco  ToolType = "falco"
	ToolLitmus ToolType = "litmus"
)

type InstallMethod string

const (
	HelmInstall     InstallMethod = "helm"
	ManifestInstall InstallMethod = "manifest"
)

type HelmInstallSpec struct {
	Repo    string                 `json:"repo"`
	Chart   string                 `json:"chart"`
	Version string                 `json:"version"`
	Values  map[string]interface{} `json:"values,omitempty"`
}

// ToolSpec defines the desired state of Tool
type ToolSpec struct {
	Type          ToolType         `json:"type"`
	Enabled       bool             `json:"enabled"`
	Namespace     string           `json:"namespace"`
	InstallMethod InstallMethod    `json:"installMethod"`
	Helm          *HelmInstallSpec `json:"helm,omitempty"`
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

	Spec   ToolSpec   `json:"spec"`
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
