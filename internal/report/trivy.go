package report

import (
	"context"
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TrivyReportHandler handles reconciling Trivy-based PolicyReports
type TrivyReportHandler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// Reconcile processes the reconciling logic for a Trivy policy report
func (h *TrivyReportHandler) Reconcile(ctx context.Context, policyReport *eirenyx.PolicyReport) error {
	// Implement the reconciliation logic specific to Trivy reports
	fmt.Printf("Reconciling Trivy report: %s\n", policyReport.Name)

	// Update the status for Trivy
	policyReport.Status.Phase = eirenyx.ReportCompleted
	return nil
}
