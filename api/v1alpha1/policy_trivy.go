package v1alpha1

type TrivyPolicySpec struct {
	Scans []TrivyScan `json:"scans"`
}

type TrivyScan struct {
	Name               string   `json:"name"`
	Image              string   `json:"image"`
	Severity           string   `json:"severity,omitempty"`
	IgnoreUnfixed      bool     `json:"ignoreUnfixed,omitempty"`
	VulnerabilityTypes []string `json:"vulnerabilityTypes,omitempty"`
	ExitCode           *int32   `json:"exitCode,omitempty"`
}
