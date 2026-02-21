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
)

// PolicyReportReconciler reconciles a PolicyReport object
type PolicyReportReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

	log.Info("Policy Report Details: ", "policyReport", policyReport.Name)
	var policy eirenyx.Policy
	err = r.Get(ctx, client.ObjectKey{
		Namespace: policyReport.Namespace,
		Name:      policyReport.Spec.PolicyRef.Name,
	}, &policy)

	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Error(err, "Policy not found", "policyName", policyReport.Spec.PolicyRef.Name)

			if deleteErr := r.Delete(ctx, &policyReport); deleteErr != nil {
				log.Error(deleteErr, "Failed to delete PolicyReport", "policyReport", policyReport.Name)
				return CompleteWithError(deleteErr)
			}
			log.Info("Deleted PolicyReport due to missing Policy", "policyReport", policyReport.Name)
			return Requeue(time.Minute * 5)
		}
		return CompleteWithError(err)
	}

	handler, err := NewReportEngine(&policyReport, Dependencies{
		Client: r.Client,
		Scheme: r.Scheme,
	})
	if err != nil {
		return CompleteWithError(err)
	}

	if err := handler.Reconcile(ctx, &policyReport); err != nil {
		log.Error(err, "Failed to reconcile policy report")
		return Requeue(time.Second * 5) // Retry after 5 seconds if there's an error
	}

	log.Info("Finished PolicyReport Reconciliation", "policyReport", policyReport.Name)
	time.Sleep(time.Minute * 5)
	return Complete()
}

// SetupWithManager sets up the controller with the Manager.
func (r *PolicyReportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eirenyx.PolicyReport{}).
		Named("policyreport").
		Complete(r)
}
