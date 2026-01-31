// +kubebuilder:object:generate=true
// +groupName=litmuschaos.io
package litmus

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ChaosResultSpec defines the desired state of ChaosResult
// The chaosresult holds the status of a chaos experiment that is listed as an item
// in the chaos engine to be run against a given app.
type ChaosResultSpec struct {
	// EngineName defines the name of chaosEngine
	EngineName string `json:"engine,omitempty"`
	// ExperimentName defines the name of chaosexperiment
	ExperimentName string `json:"experiment"`
	// InstanceID defines the instance id
	InstanceID string `json:"instance,omitempty"`
}

// ResultPhase is typecasted to string for supporting the values below.
type ResultPhase string

const (
	// ResultPhaseRunning is phase of chaosresult which is in running state
	ResultPhaseRunning ResultPhase = "Running"
	// ResultPhaseCompleted is phase of chaosresult which is in completed state
	ResultPhaseCompleted ResultPhase = "Completed"
	// ResultPhaseStopped is phase of chaosresult which is in stopped state
	ResultPhaseStopped ResultPhase = "Stopped"
)

// ResultVerdict is typecasted to string for supporting the values below.
type ResultVerdict string

const (
	// ResultVerdictPassed is verdict of chaosresult when experiment passed
	ResultVerdictPassed ResultVerdict = "Pass"
	// ResultVerdictFailed is verdict of chaosresult when experiment failed
	ResultVerdictFailed ResultVerdict = "Fail"
	// ResultVerdictStopped is verdict of chaosresult when experiment aborted
	ResultVerdictStopped ResultVerdict = "Stopped"
	// ResultVerdictAwaited is verdict of chaosresult when experiment is running
	ResultVerdictAwaited ResultVerdict = "Awaited"
)

type ProbeVerdict string

const (
	ProbeVerdictPassed  ProbeVerdict = "Passed"
	ProbeVerdictFailed  ProbeVerdict = "Failed"
	ProbeVerdictNA      ProbeVerdict = "N/A"
	ProbeVerdictAwaited ProbeVerdict = "Awaited"
)

// ChaosResultStatus defines the observed state of ChaosResult
type ChaosResultStatus struct {
	// ExperimentStatus contains the status,verdict of the experiment
	ExperimentStatus TestStatus `json:"experimentStatus"`
	// ProbeStatus contains the status of the probe
	ProbeStatuses []ProbeStatuses `json:"probeStatuses,omitempty"`
	// History contains cumulative values of verdicts
	History HistoryDetails `json:"history,omitempty"`
}

// HistoryDetails contains cumulative values of verdicts
type HistoryDetails struct {
	PassedRuns  int             `json:"passedRuns"`
	FailedRuns  int             `json:"failedRuns"`
	StoppedRuns int             `json:"stoppedRuns"`
	Targets     []TargetDetails `json:"targets,omitempty"`
}

// TargetDetails contains target details for the experiment and the chaos status
type TargetDetails struct {
	Name        string `json:"name,omitempty"`
	Kind        string `json:"kind,omitempty"`
	ChaosStatus string `json:"chaosStatus,omitempty"`
}

// ProbeStatus defines information about the status and result of the probes
type ProbeStatuses struct {
	// Name defines the name of probe
	Name string `json:"name,omitempty"`
	// Type defined the type of probe, supported values: K8sProbe, HttpProbe, CmdProbe
	Type string `json:"type,omitempty"`
	// Mode defined the mode of probe, supported values: SOT, EOT, Edge, OnChaos, Continuous
	Mode string `json:"mode,omitempty"`
	// Status defines whether a probe is pass or fail
	Status ProbeStatus `json:"status,omitempty"`
}

// ProbeStatus defines information about the status and result of the probes
type ProbeStatus struct {
	// Verdict defines the verdict of the probe, range: Passed, Failed, N/A
	Verdict ProbeVerdict `json:"verdict,omitempty"`
	// Description defines the description of probe status
	Description string `json:"description,omitempty"`
}

// TestStatus defines information about the status and results of a chaos experiment
type TestStatus struct {
	// Phase defines whether an experiment is running or completed
	Phase ResultPhase `json:"phase"`
	// Verdict defines whether an experiment result is pass or fail
	Verdict ResultVerdict `json:"verdict"`
	// FailStep defines step where the experiments fails
	FailStep string `json:"failStep,omitempty"`
	// ProbeSuccessPercentage defines the score of the probes
	ProbeSuccessPercentage string `json:"probeSuccessPercentage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +genclient
// +resource:path=chaosresult

// ChaosResult is the Schema for the chaosresults API
type ChaosResult struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChaosResultSpec   `json:"spec,omitempty"`
	Status ChaosResultStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ChaosResultList contains a list of ChaosResult
type ChaosResultList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChaosResult `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ChaosResult{}, &ChaosResultList{})
}
