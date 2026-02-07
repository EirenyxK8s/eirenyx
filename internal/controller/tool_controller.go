package controller

import (
	"context"
	"fmt"
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

func (r *ToolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Starting Tool Reconciliation")

	var tool eirenyx.Tool
	if err := r.Get(ctx, req.NamespacedName, &tool); err != nil {
		return CompleteWithError(client.IgnoreNotFound(err))
	}

	if !tool.DeletionTimestamp.IsZero() {
		log.Info("Tool is being deleted", "tool", tool.Name)

		service, ok := r.Service[tool.Spec.Type]
		if ok {
			if err := service.EnsureUninstalled(ctx, &tool); err != nil {
				log.Error(err, "Failed to uninstall tool during deletion")
				return Requeue(time.Second * 10)
			}
		}

		controllerutil.RemoveFinalizer(&tool, eirenyx.ToolFinalizer)
		if err := r.Update(ctx, &tool); err != nil {
			return CompleteWithError(err)
		}

		log.Info("Finalizer removed, deletion completed", "tool", tool.Name)
		return Complete()
	}

	if !controllerutil.ContainsFinalizer(&tool, eirenyx.ToolFinalizer) {
		controllerutil.AddFinalizer(&tool, eirenyx.ToolFinalizer)
		if err := r.Update(ctx, &tool); err != nil {
			return CompleteWithError(err)
		}
	}

	log.Info("Reconciling Tool", "tool", tool.Spec.Type, "enabled", tool.Spec.Enabled)

	service, ok := r.Service[tool.Spec.Type]
	if !ok {
		return CompleteWithError(fmt.Errorf("tool type %q not found", tool.Spec.Type))
	}

	if tool.Spec.Enabled {
		if err := service.EnsureInstalled(ctx, &tool); err != nil {
			log.Error(err, "Failed to install tool")
			return Requeue(time.Second * 5)
		}
	} else {
		if err := service.EnsureUninstalled(ctx, &tool); err != nil {
			log.Error(err, "Failed to uninstall tool")
			return Requeue(time.Second * 5)
		}
	}

	healthy := service.CheckHealth(ctx, &tool)
	tool.Status.Installed = tool.Spec.Enabled
	tool.Status.Healthy = healthy
	_ = r.Status().Update(ctx, &tool)

	log.Info("Finished Tool Reconciliation")
	return Complete()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ToolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eirenyx.Tool{}).
		Named("tool").
		Complete(r)
}
