package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// PolicySpec defines the desired state of Policy
type PolicySpec struct {
	Base BasePolicySpec `json:",inline"`

	Falco *FalcoPolicySpec `json:"falco,omitempty"`
}

// PolicyStatus defines the observed state of the Policy.
type PolicyStatus struct {
	Base BasePolicyStatus `json:",inline"`
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
