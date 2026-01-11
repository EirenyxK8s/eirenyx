package controller

import (
	"context"
	"time"

	"github.com/EirenyxK8s/eirenyx/internal/tools"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

// ToolReconciler reconciles a Tool object
type ToolReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Service map[eirenyx.ToolType]tools.ToolService
}

// -----------------------------------------------------------------------------
// Eirenyx CRDs
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=tools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=tools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=tools/finalizers,verbs=update

// -----------------------------------------------------------------------------
// Namespaces (cluster-scoped)
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create

// -----------------------------------------------------------------------------
// Core Kubernetes resources (used by installed tools)
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups="",resources=serviceaccounts;configmaps;secrets;services,verbs=*
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;statefulsets;replicasets;pods,verbs=*

// -----------------------------------------------------------------------------
// RBAC (created by Helm charts)
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings;clusterroles;clusterrolebindings,verbs=*

// -----------------------------------------------------------------------------
// CRDs (installed by tools like Trivy Operator)
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete

// -----------------------------------------------------------------------------
// Trivy Operator CRs (cluster-scoped & namespaced)
// Required for Helm install, upgrade, and uninstall
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=aquasecurity.github.io,resources=*,verbs=*

func (r *ToolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Starting Tool Reconciliation")

	var tool eirenyx.Tool
	if err := r.Get(ctx, req.NamespacedName, &tool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !tool.DeletionTimestamp.IsZero() {
		log.Info("Tool is being deleted", "tool", tool.Name)

		service, ok := r.Service[tool.Spec.Type]
		if ok {
			if err := service.EnsureUninstalled(ctx, &tool); err != nil {
				log.Error(err, "Failed to uninstall tool during deletion")
				return ctrl.Result{RequeueAfter: time.Second * 10}, err
			}
		}

		controllerutil.RemoveFinalizer(&tool, eirenyx.ToolFinalizer)
		if err := r.Update(ctx, &tool); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Finalizer removed, deletion completed", "tool", tool.Name)
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(&tool, eirenyx.ToolFinalizer) {
		controllerutil.AddFinalizer(&tool, eirenyx.ToolFinalizer)
		if err := r.Update(ctx, &tool); err != nil {
			return ctrl.Result{}, err
		}
	}

	log.Info("Reconciling Tool", "tool", tool.Spec.Type, "enabled", tool.Spec.Enabled)

	service, ok := r.Service[tool.Spec.Type]
	if !ok {
		log.Error(nil, "Tool type not found", "tool", tool.Spec.Type)
		return ctrl.Result{}, nil
	}

	if tool.Spec.Enabled {
		if err := service.EnsureInstalled(ctx, &tool); err != nil {
			return ctrl.Result{RequeueAfter: time.Second}, err
		}
	} else {
		if err := service.EnsureUninstalled(ctx, &tool); err != nil {
			return ctrl.Result{RequeueAfter: time.Second}, err
		}
	}

	healthy := service.CheckHealth(ctx, &tool)
	tool.Status.Installed = tool.Spec.Enabled
	tool.Status.Healthy = healthy
	_ = r.Status().Update(ctx, &tool)

	log.Info("Finished Tool Reconciliation")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ToolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eirenyx.Tool{}).
		Named("tool").
		Complete(r)
}
