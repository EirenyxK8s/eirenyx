// +kubebuilder:object:generate=true
// +groupName=litmuschaos.io
package litmus

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ChaosEngineSpec struct {
	AppInfo     ApplicationParams `json:"appinfo,omitempty"`
	EngineState EngineState       `json:"engineState"`
	Experiments []ExperimentList  `json:"experiments"`
}

type EngineState string

const (
	EngineStateActive EngineState = "active"
	EngineStateStop   EngineState = "stop"
)

type ExperimentStatus string

const (
	ExperimentStatusRunning   ExperimentStatus = "Running"
	ExperimentStatusCompleted ExperimentStatus = "Completed"
	ExperimentStatusWaiting   ExperimentStatus = "Waiting for Job Creation"
	ExperimentStatusNotFound  ExperimentStatus = "ChaosExperiment Not Found"
	ExperimentStatusAborted   ExperimentStatus = "Forcefully Aborted"
	ExperimentSkipped         ExperimentStatus = "Skipped"
)

type EngineStatus string

const (
	EngineStatusInitialized EngineStatus = "initialized"
	EngineStatusCompleted   EngineStatus = "completed"
	EngineStatusStopped     EngineStatus = "stopped"
)

type ChaosEngineStatus struct {
	EngineStatus EngineStatus         `json:"engineStatus"`
	Experiments  []ExperimentStatuses `json:"experiments"`
}

type ApplicationParams struct {
	Appns    string `json:"appns,omitempty"`
	Applabel string `json:"applabel,omitempty"`
	AppKind  string `json:"appkind,omitempty"`
}

type ExperimentList struct {
	Name string               `json:"name"`
	Spec ExperimentAttributes `json:"spec"`
}

// ExperimentAttributes defines attributes of experiments
type ExperimentAttributes struct {
	Components ExperimentComponents `json:"components,omitempty"`
}

type ExperimentComponents struct {
	ENV                        []corev1.EnvVar               `json:"env,omitempty"`
	ConfigMaps                 []ConfigMap                   `json:"configMaps,omitempty"`
	Secrets                    []Secret                      `json:"secrets,omitempty"`
	ExperimentAnnotations      map[string]string             `json:"experimentAnnotations,omitempty"`
	ExperimentImage            string                        `json:"experimentImage,omitempty"`
	ExperimentImagePullSecrets []corev1.LocalObjectReference `json:"experimentImagePullSecrets,omitempty"`
	NodeSelector               map[string]string             `json:"nodeSelector,omitempty"`
	StatusCheckTimeouts        StatusCheckTimeout            `json:"statusCheckTimeouts,omitempty"`
	Resources                  corev1.ResourceRequirements   `json:"resources,omitempty"`
	Tolerations                []corev1.Toleration           `json:"tolerations,omitempty"`
}

type StatusCheckTimeout struct {
	Delay   int `json:"delay,omitempty"`
	Timeout int `json:"timeout,omitempty"`
}

type ExperimentStatuses struct {
	Name           string           `json:"name"`
	Runner         string           `json:"runner"`
	ExpPod         string           `json:"experimentPod"`
	Status         ExperimentStatus `json:"status"`
	Verdict        string           `json:"verdict"`
	LastUpdateTime metav1.Time      `json:"lastUpdateTime"`
}

// +genclient
// +resource:path=chaosengine
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ChaosEngine is the Schema for the chaosengines API
type ChaosEngine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChaosEngineSpec   `json:"spec,omitempty"`
	Status ChaosEngineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ChaosEngineList contains a list of ChaosEngine
type ChaosEngineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChaosEngine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ChaosEngine{}, &ChaosEngineList{})
}
