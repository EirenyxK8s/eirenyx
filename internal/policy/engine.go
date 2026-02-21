package policy

import (
	"context"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

type Engine interface {
	Validate(policy *eirenyx.Policy) error
	Reconcile(ctx context.Context, policy *eirenyx.Policy) error
	Cleanup(ctx context.Context, policy *eirenyx.Policy) error
	GenerateReport(ctx context.Context, policy *eirenyx.Policy) (*eirenyx.PolicyReport, error)
}
