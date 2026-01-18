package policy

import (
	"errors"
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/policy/falco"
	"github.com/EirenyxK8s/eirenyx/internal/policy/trivy"
)

func NewEngine(policy *eirenyx.Policy, deps Dependencies) (Engine, error) {
	switch policy.Spec.Base.Type {
	case eirenyx.PolicyTypeFalco:
		return &falco.Engine{
			Client: deps.Client,
			Scheme: deps.Scheme,
		}, nil
	case eirenyx.PolicyTypeTrivy:
		return &trivy.Engine{
			Client: deps.Client,
			Scheme: deps.Scheme,
		}, nil
	default:
		return nil, errors.New(fmt.Sprintf("unsupported policy type: %s", policy.Spec.Base.Type))
	}
}
