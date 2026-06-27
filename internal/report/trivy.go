package report

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/client/k8s"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// trivyContainerName matches the container name set by the Trivy engine
// when it creates the scan Job. See internal/policy/trivy/engine.go.
const trivyContainerName = "trivy"

// TrivyReportHandler builds a PolicyReport by reading the JSON output of the
// scan Jobs created by the Trivy policy engine. We deliberately do NOT consult
// trivy-operator's VulnerabilityReport CRDs — those describe whatever container
// images run in the cluster (including our own scan-Job pods, which scan the
// scanner image itself); they don't reflect the image the user asked us to scan.
type TrivyReportHandler struct {
	Client    client.Client
	Scheme    *runtime.Scheme
	K8sClient *k8s.Client
}

// Reconcile is called by the PolicyReport controller. It returns an error when
// the scan is not yet finished — the controller's RequeueAfter path picks it
// up again a few seconds later (see internal/controller/policyreport_controller.go).
func (h *TrivyReportHandler) Reconcile(ctx context.Context, policyReport *eirenyx.PolicyReport) error {
	var policy eirenyx.Policy
	if err := h.Client.Get(ctx, client.ObjectKey{
		Name:      policyReport.Spec.PolicyRef.Name,
		Namespace: policyReport.Namespace,
	}, &policy); err != nil {
		return err
	}

	if policy.Spec.Trivy == nil || len(policy.Spec.Trivy.Scans) == 0 {
		return fmt.Errorf("policy %s/%s has no trivy scans", policy.Namespace, policy.Name)
	}

	// Collect results across every scan in the policy. We aggregate
	// vulnerabilities into a single details payload — the UI flattens them too.
	var (
		allVulns       []UIVulnerability
		firstImage     string
		completedScans int
		pendingScans   []string
	)

	for _, scan := range policy.Spec.Trivy.Scans {
		jobName := getScanJobName(policy.Name, scan.Name)

		var job batchv1.Job
		if err := h.Client.Get(ctx, client.ObjectKey{
			Name:      jobName,
			Namespace: policy.Namespace,
		}, &job); err != nil {
			if apierrors.IsNotFound(err) {
				// Engine hasn't created the Job yet — keep Pending and retry.
				pendingScans = append(pendingScans, scan.Name)
				continue
			}
			return fmt.Errorf("getting scan job %s/%s: %w", policy.Namespace, jobName, err)
		}

		if !jobFinished(&job) {
			pendingScans = append(pendingScans, scan.Name)
			continue
		}

		// Job finished — find the scan pod and read its logs.
		report, err := h.readScanReport(ctx, policy.Namespace, jobName)
		if err != nil {
			return fmt.Errorf("scan %q: %w", scan.Name, err)
		}

		// The user wrote the image string in the policy spec. Prefer it over
		// Trivy's parsed ArtifactName so the UI shows exactly what was requested
		// (Trivy may normalize "nginx:1.31-perl" to "docker.io/library/nginx:1.31-perl").
		if firstImage == "" {
			firstImage = scan.Image
			if firstImage == "" {
				firstImage = report.ArtifactName
			}
		}

		allVulns = append(allVulns, flattenVulnerabilities(report)...)
		completedScans++
	}

	// If anything is still in flight, mark Running and return an error so the
	// controller requeues. This mirrors the previous handler's contract — see
	// internal/controller/policyreport_controller.go for the requeue path.
	if len(pendingScans) > 0 {
		policyReport.Status.Phase = eirenyx.ReportRunning
		if err := h.Client.Status().Update(ctx, policyReport); err != nil {
			return err
		}
		return fmt.Errorf("trivy scans still pending: %s", strings.Join(pendingScans, ","))
	}

	// Final summary. For a vulnerability scan, "totalChecks" = number of CVEs
	// found, and every CVE is a failed check. There is no notion of a "passed
	// check" for trivy — a passing scan is one with zero findings. The earlier
	// implementation counted Medium/Low CVEs as "passed", which is meaningless.
	total := int32(len(allVulns))
	failed := total
	verdict := eirenyx.VerdictPass
	if failed > 0 {
		verdict = eirenyx.VerdictFail
	}

	detailsBytes, err := json.Marshal(TrivyDetails{
		Image:           firstImage,
		Vulnerabilities: allVulns,
		ReportCount:     completedScans,
	})
	if err != nil {
		return err
	}

	policyReport.Status.Phase = eirenyx.ReportCompleted
	policyReport.Status.Summary = eirenyx.ReportSummary{
		Verdict:     verdict,
		TotalChecks: total,
		Passed:      0,
		Failed:      failed,
	}
	policyReport.Status.Details = runtime.RawExtension{Raw: detailsBytes}

	return h.Client.Status().Update(ctx, policyReport)
}

// readScanReport locates the Pod that the Job ran and parses its stdout as a
// Trivy JSON report. Returns an error the caller propagates so the controller
// requeues — except in the "no pod yet" case, where we treat it as "still
// pending" by returning a retry-shaped error.
func (h *TrivyReportHandler) readScanReport(ctx context.Context, namespace, jobName string) (*TrivyCLIReport, error) {
	pods, err := h.K8sClient.ListPods(ctx, namespace, map[string]string{"job-name": jobName})
	if err != nil {
		return nil, err
	}
	if len(pods) == 0 {
		return nil, fmt.Errorf("no pods found for job %s yet", jobName)
	}

	// Pick the most recently completed pod. Job pods generally don't restart
	// (RestartPolicy=Never on the trivy Job), so usually len(pods)==1 — but be
	// robust in case the user re-applied the policy.
	pod := pickNewestPod(pods)
	if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
		return nil, fmt.Errorf("scan pod %s/%s phase=%s, not yet terminal", pod.Namespace, pod.Name, pod.Status.Phase)
	}

	logs, err := h.K8sClient.GetPodLogs(ctx, pod.Namespace, pod.Name, trivyContainerName)
	if err != nil {
		return nil, fmt.Errorf("reading scan pod logs: %w", err)
	}
	logs = strings.TrimSpace(logs)
	if logs == "" {
		return nil, fmt.Errorf("scan pod %s/%s produced no output", pod.Namespace, pod.Name)
	}

	// Trivy prints a JSON object — even on "no vulns found" cases, with empty
	// Results. We don't strip non-JSON prefix lines: --quiet suppresses progress
	// output, so the buffer should be valid JSON in its entirety. If Trivy
	// happens to print warnings to stdout before the JSON, the unmarshal will
	// fail and we surface the underlying error.
	var report TrivyCLIReport
	if err := json.Unmarshal([]byte(logs), &report); err != nil {
		// Be helpful: include the first chunk of the pod output in the error so
		// operators can see what came back instead of valid JSON.
		preview := logs
		if len(preview) > 200 {
			preview = preview[:200] + "…"
		}
		return nil, fmt.Errorf("parsing trivy json from pod %s/%s: %w (got: %q)", pod.Namespace, pod.Name, err, preview)
	}
	return &report, nil
}

// jobFinished reports whether a Job has reached a terminal state. We treat
// Succeeded and Failed counts as the signal; controller-runtime's batchv1 Job
// also has conditions, but the count is more reliable across versions.
func jobFinished(job *batchv1.Job) bool {
	return job.Status.Succeeded >= 1 || job.Status.Failed >= 1
}

// pickNewestPod returns the pod with the newest CreationTimestamp. We need
// this in case the scan-job was re-created (e.g. policy generation bumped)
// and lingering old pods are still in the namespace.
func pickNewestPod(pods []corev1.Pod) corev1.Pod {
	newest := pods[0]
	for _, p := range pods[1:] {
		if p.CreationTimestamp.After(newest.CreationTimestamp.Time) {
			newest = p
		}
	}
	return newest
}

// getScanJobName mirrors the formatter in internal/policy/trivy/engine.go so
// both sides agree on the Job name. Kept private here to avoid an awkward
// import cycle; if either side changes, update both.
func getScanJobName(policyName, scanName string) string {
	return fmt.Sprintf("eirenyx-trivy-%s-%s", policyName, scanName)
}
