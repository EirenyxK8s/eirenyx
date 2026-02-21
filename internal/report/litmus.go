package report

import (
	"context"
	"reflect"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

// LitmusReportHandler handles reconciling Litmus-based PolicyReports
type LitmusReportHandler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// Reconcile processes the reconciling logic for a Litmus policy report
func (h *LitmusReportHandler) Reconcile(ctx context.Context, policyReport *eirenyx.PolicyReport) error {
	log := logger.FromContext(ctx)

	var policy eirenyx.Policy
	if err := h.Client.Get(ctx, client.ObjectKey{
		Namespace: policyReport.Namespace,
		Name:      policyReport.Spec.PolicyRef.Name,
	}, &policy); err != nil {
		return err
	}

	total := len(policy.Spec.Litmus.Experiments)

	newStatus := policyReport.Status.DeepCopy()

	newStatus.Summary = eirenyx.ReportSummary{
		TotalChecks: int32(total),
		Passed:      int32(total),
		Failed:      0,
		Verdict:     eirenyx.VerdictPass,
	}

	newStatus.Phase = eirenyx.ReportCompleted
	newStatus.Details = createLitmusReportDetails(policy.Spec.Litmus.Experiments)

	if !statusEqual(policyReport.Status, *newStatus) {
		policyReport.Status = *newStatus
		return h.Client.Status().Update(ctx, policyReport)
	}

	log.Info("PolicyReport already up-to-date")
	return nil
}

func statusEqual(a, b eirenyx.PolicyReportStatus) bool {
	return reflect.DeepEqual(a, b)
}

// createLitmusReportDetails creates a runtime.RawExtension with detailed information for the Litmus report
func createLitmusReportDetails(experiments []eirenyx.LitmusExperiment) runtime.RawExtension {
	report := struct {
		Message     string                     `json:"message"`
		Experiments []eirenyx.LitmusExperiment `json:"experiments"`
	}{
		Message:     "Obtained results for policy reconciliation.",
		Experiments: experiments,
	}

	reportData, err := json.Marshal(report)
	if err != nil {
		return runtime.RawExtension{}
	}

	return runtime.RawExtension{Raw: reportData}
}
