package v1alpha1

type LitmusPolicySpec struct {
	Experiments []LitmusExperiment `json:"experiments"`
}

type LitmusExperiment struct {
	Name string `json:"name"`

	ExperimentRef string `json:"experimentRef"`
	EngineRef     string `json:"engineRef,omitempty"`

	TargetNamespace string `json:"targetNamespace,omitempty"`

	AppInfo LitmusAppInfo `json:"appInfo"`

	Mode       string            `json:"mode,omitempty"`
	Duration   string            `json:"duration,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`

	ExpectedResult LitmusExpectedResult `json:"expectedResult,omitempty"`
}

type LitmusAppInfo struct {
	AppNamespace string `json:"appNamespace"`
	AppLabel     string `json:"appLabel"`
	AppKind      string `json:"appKind"`
}

type LitmusExpectedResult struct {
	Verdict          string `json:"verdict,omitempty"`
	FailOnChaosError *bool  `json:"failOnChaosError,omitempty"`
}
