package report

import (
	"context"
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LitmusReportHandler handles reconciling Litmus-based PolicyReports
type LitmusReportHandler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// Reconcile processes the reconciling logic for a Litmus policy report
func (h *LitmusReportHandler) Reconcile(ctx context.Context, policyReport *eirenyx.PolicyReport) error {
	// Implement the reconciliation logic specific to Litmus reports
	fmt.Printf("Reconciling Litmus report: %s\n", policyReport.Name)

	// Update the status for Litmus
	policyReport.Status.Phase = eirenyx.ReportCompleted
	return nil
}
