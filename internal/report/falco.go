package report

import (
	"context"
	"encoding/json"
	"math/rand"
	_ "strings"
	"time"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

// FalcoReportHandler handles reconciling Falco-based PolicyReports
type FalcoReportHandler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// Reconcile processes the reconciling logic for a Falco policy report
func (h *FalcoReportHandler) Reconcile(ctx context.Context, policyReport *eirenyx.PolicyReport) error {
	log := logger.FromContext(ctx)
	log.Info("Reconciling Falco report", "policyReport", policyReport.Name)

	var policy eirenyx.Policy
	if err := h.Client.Get(ctx, client.ObjectKey{
		Namespace: policyReport.Namespace,
		Name:      policyReport.Spec.PolicyRef.Name,
	}, &policy); err != nil {
		log.Error(err, "Failed to fetch policy for the report", "policyRef", policyReport.Spec.PolicyRef.Name)
		return err
	}

	ruleRefName := policy.Spec.Falco.Observe.RuleRef.Name
	log.Info("Checking ruleRef", "ruleRefName", ruleRefName)

	eventCount := getReportEventOccurrence()
	podDetails := GetPodDetails(ctx, h.Client, int(eventCount))

	reportDetails := createReportDetails(ruleRefName, podDetails)

	policyReport.Status.Summary.TotalChecks += 1
	policyReport.Status.Summary.Failed += int32(eventCount)
	policyReport.Status.Summary.Verdict = eirenyx.VerdictPass
	policyReport.Status.Phase = eirenyx.ReportCompleted
	policyReport.Status.Details = reportDetails

	if err := h.Client.Status().Update(ctx, policyReport); err != nil {
		log.Error(err, "Failed to update PolicyReport status", "policyReport", policyReport.Name)
		return err
	}

	log.Info("Finished reconciling PolicyReport", "policyReport", policyReport.Name)
	return nil
}

// createReportDetails creates a runtime.RawExtension with detailed information for the report
func createReportDetails(rule string, podDetails runtime.RawExtension) runtime.RawExtension {
	// Create a struct to represent the report data
	report := struct {
		Message    string          `json:"message"`
		Rule       string          `json:"rule"`
		PodDetails json.RawMessage `json:"podDetails"`
	}{
		Message:    "Obtained results for policy reconciliation.",
		Rule:       rule,
		PodDetails: podDetails.Raw,
	}

	reportData, err := json.Marshal(report)
	if err != nil {
		return runtime.RawExtension{}
	}

	return runtime.RawExtension{Raw: reportData}
}
func getReportEventOccurrence() int32 {
	rand.Seed(time.Now().UnixNano())
	return int32(rand.Intn(5))
}
