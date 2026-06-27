package controller

import (
	"context"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/client/k8s"
)

// PolicyReportReconciler reconciles a PolicyReport object
type PolicyReportReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// K8sClient is required by report handlers that need to talk to the core
	// Kubernetes API directly (e.g. Trivy reads pod logs to extract scan output).
	K8sClient *k8s.Client
}

func (r *PolicyReportReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciling PolicyReport")

	var policyReport eirenyx.PolicyReport
	err := r.Get(ctx, req.NamespacedName, &policyReport)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("PolicyReport not found, ignoring", "policyReport", req.NamespacedName)
			return Complete()
		}
		return CompleteWithError(client.IgnoreNotFound(err))
	}

	if policyReport.Status.Phase == eirenyx.ReportCompleted {
		log.Info("PolicyReport already completed, skipping", "policyReport", policyReport.Name)
		return Complete()
	}

	log.Info("Reconciling PolicyReport", "policyReport", policyReport.Name, "phase", policyReport.Status.Phase)

	var policy eirenyx.Policy
	err = r.Get(ctx, client.ObjectKey{
		Namespace: policyReport.Namespace,
		Name:      policyReport.Spec.PolicyRef.Name,
	}, &policy)

	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Error(err, "Policy not found, deleting orphaned report", "policyName", policyReport.Spec.PolicyRef.Name)
			if deleteErr := r.Delete(ctx, &policyReport); deleteErr != nil {
				log.Error(deleteErr, "Failed to delete orphaned PolicyReport", "policyReport", policyReport.Name)
				return CompleteWithError(deleteErr)
			}
			log.Info("Deleted orphaned PolicyReport", "policyReport", policyReport.Name)
			return Complete()
		}
		return CompleteWithError(err)
	}

	handler, err := NewReportEngine(&policyReport, Dependencies{
		Client:    r.Client,
		Scheme:    r.Scheme,
		K8sClient: r.K8sClient,
	})
	if err != nil {
		return CompleteWithError(err)
	}

	if err := handler.Reconcile(ctx, &policyReport); err != nil {
		log.Error(err, "Failed to reconcile policy report, requeueing")
		return RequeueAfter(time.Second * 5)
	}

	log.Info("Finished PolicyReport Reconciliation", "policyReport", policyReport.Name)
	return Complete()
}

// SetupWithManager sets up the controller with the Manager.
func (r *PolicyReportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eirenyx.PolicyReport{}).
		Named("policyreport").
		Complete(r)
}
