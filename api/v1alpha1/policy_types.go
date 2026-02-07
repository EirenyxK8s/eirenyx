package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const PolicyFinalizer = "eirenyx.policy/finalizer"

// PolicySpec defines the desired state of Policy
type PolicySpec struct {
	Type    PolicyType   `json:"type"`
	Enabled bool         `json:"enabled"`
	Target  PolicyTarget `json:"target,omitempty"`

	Falco  *FalcoPolicySpec  `json:"falco,omitempty"`
	Trivy  *TrivyPolicySpec  `json:"trivy,omitempty"`
	Litmus *LitmusPolicySpec `json:"litmus,omitempty"`
}

// PolicyStatus defines the observed state of the Policy.
type PolicyStatus struct {
	Phase string `json:"phase,omitempty"`
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions  []metav1.Condition `json:"conditions,omitempty"`
	LastReport  string             `json:"lastReport,omitempty"`
	ObservedGen int64              `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Policy is the Schema for the policies API
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitzero"`

	Spec PolicySpec `json:"spec"`
	// +optional
	Status PolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Policy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Policy{}, &PolicyList{})
}
