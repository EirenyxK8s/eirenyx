package v1alpha1

type FalcoPolicySpec struct {
	Rules []FalcoRule `json:"rules"`
}

type FalcoRule struct {
	Name      string   `json:"name"`
	Condition string   `json:"condition"`
	Output    string   `json:"output"`
	Priority  string   `json:"priority"`
	Tags      []string `json:"tags,omitempty"`
}
