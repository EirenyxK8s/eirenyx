package report

import (
	"context"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

// Handler is an interface for handling policy reports of different types
type Handler interface {
	Reconcile(ctx context.Context, policyReport *eirenyx.PolicyReport) error
}
