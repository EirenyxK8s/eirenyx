package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ReportPhase string

const (
	ReportPending   ReportPhase = "Pending"
	ReportRunning   ReportPhase = "Running"
	ReportCompleted ReportPhase = "Completed"
	ReportFailed    ReportPhase = "Failed"
)

type Verdict string

const (
	VerdictPass    Verdict = "Pass"
	VerdictFail    Verdict = "Fail"
	VerdictUnknown Verdict = "Unknown"
)

type ReportSummary struct {
	Verdict     Verdict `json:"verdict,omitempty"`
	TotalChecks int32   `json:"totalChecks,omitempty"`
	Passed      int32   `json:"passed,omitempty"`
	Failed      int32   `json:"failed,omitempty"`
}

type PolicyReference struct {
	Name       string `json:"name"`
	Generation int64  `json:"generation"`
}

// PolicyReportSpec defines the desired state of PolicyReport
type PolicyReportSpec struct {
	PolicyRef PolicyReference `json:"policyRef"`
	Type      PolicyType      `json:"type"`
}

// PolicyReportStatus defines the observed state of the PolicyReport.
type PolicyReportStatus struct {
	Phase   ReportPhase          `json:"phase,omitempty"`
	Summary ReportSummary        `json:"summary,omitempty"`
	Details runtime.RawExtension `json:"details,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PolicyReport is the Schema for the policyreports API
type PolicyReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitzero"`

	Spec PolicyReportSpec `json:"spec"`
	// +optional
	Status PolicyReportStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PolicyReportList contains a list of PolicyReport
type PolicyReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PolicyReport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PolicyReport{}, &PolicyReportList{})
}
