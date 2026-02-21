package v1alpha1

const (
	FalcoRulesConfigMapName = "eirenyx-falco-rules"
	FalcoRulesKey           = "eirenyx.yaml"
)

type FalcoPolicySpec struct {
	Observe FalcoObserveSpec `json:"observe"`
	Report  *FalcoReportSpec `json:"report,omitempty"`
}

type FalcoObserveSpec struct {
	RuleRef      *FalcoRuleRef      `json:"ruleRef,omitempty"`
	RuleSelector *FalcoRuleSelector `json:"ruleSelector,omitempty"`
}

type FalcoRuleRef struct {
	Name string `json:"name"`
}

type FalcoRuleSelector struct {
	Tags       []string `json:"tags,omitempty"`
	Priorities []string `json:"priorities,omitempty"`
}

type FalcoReportSpec struct {
	Create            bool   `json:"create"`
	Severity          string `json:"severity,omitempty"`
	AggregationWindow string `json:"aggregationWindow,omitempty"`
}
