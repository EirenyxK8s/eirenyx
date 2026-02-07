package v1alpha1

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
