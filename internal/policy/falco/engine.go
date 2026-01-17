package falco

import (
	"context"

	eirenyxv1alpha1 "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Engine struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (e *Engine) Validate(policy *eirenyxv1alpha1.Policy) error {
	return nil
}

func (e *Engine) Reconcile(ctx context.Context, policy *eirenyxv1alpha1.Policy) error {
	return nil
}

func (e *Engine) Cleanup(ctx context.Context, policy *eirenyxv1alpha1.Policy) error {
	return nil
}

func (e *Engine) GenerateReport(ctx context.Context, policy *eirenyxv1alpha1.Policy) (string, error) {
	return "falco-report-name", nil
}
