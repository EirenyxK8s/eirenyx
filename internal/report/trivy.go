package report

import (
	"context"
	"encoding/json"
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	aquav1 "github.com/aquasecurity/trivy-operator/pkg/apis/aquasecurity/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type trivyDetails struct {
	Image           string                 `json:"image,omitempty"`
	Vulnerabilities []aquav1.Vulnerability `json:"vulnerabilities,omitempty"`
	ReportCount     int                    `json:"reportCount,omitempty"`
}

type TrivyReportHandler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (h *TrivyReportHandler) Reconcile(ctx context.Context, policyReport *eirenyx.PolicyReport) error {

	var policy eirenyx.Policy
	if err := h.Client.Get(ctx, client.ObjectKey{
		Name:      policyReport.Spec.PolicyRef.Name,
		Namespace: policyReport.Namespace,
	}, &policy,
	); err != nil {
		return err
	}

	var vulnReports aquav1.VulnerabilityReportList
	if err := h.Client.List(ctx, &vulnReports, client.InNamespace(policyReport.Namespace)); err != nil {
		return err
	}

	var relevantReports []aquav1.VulnerabilityReport
	for _, scan := range policy.Spec.Trivy.Scans {
		expectedJobName := fmt.Sprintf("eirenyx-trivy-%s-%s", policy.Name, scan.Name)
		for _, vr := range vulnReports.Items {
			if vr.Labels["trivy-operator.resource.kind"] == "Job" &&
				vr.Labels["trivy-operator.resource.name"] == expectedJobName {
				relevantReports = append(relevantReports, vr)
			}
		}
	}

	if len(relevantReports) == 0 {
		policyReport.Status.Phase = eirenyx.ReportRunning
		err := h.Client.Status().Update(ctx, policyReport)
		if err == nil {
			err = fmt.Errorf("report for policy %s is not complete, should reconcile", policyReport.Spec.PolicyRef.Name)
		}
		return err
	}

	var total int32
	var failed int32
	var allVulns []aquav1.Vulnerability

	for _, vr := range relevantReports {
		summary := vr.Report.Summary

		total += int32(
			summary.CriticalCount +
				summary.HighCount +
				summary.MediumCount +
				summary.LowCount,
		)

		if summary.CriticalCount > 0 || summary.HighCount > 0 {
			failed++
		}

		allVulns = append(allVulns, vr.Report.Vulnerabilities...)
	}

	verdict := eirenyx.VerdictPass
	if failed > 0 {
		verdict = eirenyx.VerdictFail
	}

	details := trivyDetails{
		Vulnerabilities: allVulns,
		ReportCount:     len(relevantReports),
	}

	detailsBytes, err := json.Marshal(details)
	if err != nil {
		return err
	}
	policyReport.Status.Phase = eirenyx.ReportCompleted
	policyReport.Status.Summary = eirenyx.ReportSummary{
		Verdict:     verdict,
		TotalChecks: total,
		Passed:      total - failed,
		Failed:      failed,
	}
	policyReport.Status.Details = runtime.RawExtension{
		Raw: detailsBytes,
	}

	return h.Client.Status().Update(ctx, policyReport)
}
