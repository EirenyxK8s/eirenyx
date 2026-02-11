package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	eirenyxv1alpha1 "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

// PolicyReportReconciler reconciles a PolicyReport object
type PolicyReportReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *PolicyReportReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	return Complete()
}

// SetupWithManager sets up the controller with the Manager.
func (r *PolicyReportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eirenyxv1alpha1.PolicyReport{}).
		Named("policyreport").
		Complete(r)
}
