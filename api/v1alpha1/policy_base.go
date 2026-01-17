package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type PolicyType string

const (
	PolicyTypeFalco  PolicyType = "falco"
	PolicyTypeTrivy  PolicyType = "trivy"
	PolicyTypeLitmus PolicyType = "litmus"
)

type PolicyTarget struct {
	NamespaceSelector []string `json:"namespaceSelector,omitempty"`
	NodeSelector      []string `json:"nodeSelector,omitempty"`
}

type BasePolicySpec struct {
	Type    PolicyType   `json:"type"`
	Enabled bool         `json:"enabled"`
	Target  PolicyTarget `json:"target,omitempty"`
}

type BasePolicyStatus struct {
	Phase string `json:"phase,omitempty"`
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions  []metav1.Condition `json:"conditions,omitempty"`
	LastReport  string             `json:"lastReport,omitempty"`
	ObservedGen int64              `json:"observedGeneration,omitempty"`
}
