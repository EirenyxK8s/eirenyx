package policy

import (
	"fmt"

	eirenyxv1alpha1 "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/policy/falco"
)

func NewEngine(policy *eirenyxv1alpha1.Policy, deps Dependencies) (Engine, error) {
	switch policy.Spec.Base.Type {
	case eirenyxv1alpha1.PolicyTypeFalco:
		return &falco.Engine{
			Client: deps.Client,
			Scheme: deps.Scheme,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported policy type: %s", policy.Spec.Base.Type)
	}
}
