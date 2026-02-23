package controller

import (
	"context"
	"time"

	"github.com/EirenyxK8s/eirenyx/internal/client/k8s"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Scheme    *runtime.Scheme
	K8sClient k8s.Client
}

func (r *ToolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Starting Tool Reconciliation")

	var tool eirenyx.Tool
	if err := r.Get(ctx, req.NamespacedName, &tool); err != nil {
		return CompleteWithError(client.IgnoreNotFound(err))
	}

	if !tool.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, &tool)
	}

	if !controllerutil.ContainsFinalizer(&tool, eirenyx.ToolFinalizer) {
		patch := client.MergeFrom(tool.DeepCopy())
		controllerutil.AddFinalizer(&tool, eirenyx.ToolFinalizer)
		if err := r.Patch(ctx, &tool, patch); err != nil {
			return CompleteWithError(err)
		}
	}

	svc, err := NewToolService(&tool, Dependencies{
		Client:    r.Client,
		Scheme:    r.Scheme,
		K8sClient: &r.K8sClient,
	})
	if err != nil {
		return r.setFailedCondition(ctx, &tool, "UnsupportedToolType", err.Error())
	}

	log.Info("Reconciling Tool", "service", svc.Name(), "enabled", tool.Spec.Enabled)

	if tool.Spec.Enabled {
		if err := svc.EnsureInstalled(ctx, &tool); err != nil {
			log.Error(err, "Failed to install tool")
			return RequeueAfter(5 * time.Second)
		}
	} else {
		if err := svc.EnsureUninstalled(ctx, &tool); err != nil {
			log.Error(err, "Failed to uninstall tool")
			return RequeueAfter(5 * time.Second)
		}
	}

	healthy, err := svc.CheckHealth(ctx, &tool)
	if err != nil {
		log.Error(err, "Health check failed")
		return RequeueAfter(10 * time.Second)
	}

	tool.Status.Installed = tool.Spec.Enabled
	tool.Status.Healthy = healthy
	if err := r.Status().Update(ctx, &tool); err != nil {
		log.Error(err, "Failed to update tool status")
		return RequeueAfter(5 * time.Second)
	}

	if !healthy {
		log.Info("Tool not yet healthy, requeuing", "tool", tool.Name)
		return RequeueAfter(5 * time.Second)
	}

	log.Info("Finished Tool Reconciliation", "tool", tool.Name)
	return Complete()
}

func (r *ToolReconciler) handleDeletion(ctx context.Context, tool *eirenyx.Tool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Tool is being deleted", "tool", tool.Name)

	svc, err := NewToolService(tool, Dependencies{
		Client:    r.Client,
		Scheme:    r.Scheme,
		K8sClient: &r.K8sClient,
	})
	if err != nil {
		log.Info("Unknown tool type during deletion, skipping uninstall", "type", tool.Spec.Type)
	} else {
		if err := svc.EnsureUninstalled(ctx, tool); err != nil {
			log.Error(err, "Failed to uninstall tool during deletion")
			return RequeueAfter(10 * time.Second)
		}
	}

	patch := client.MergeFrom(tool.DeepCopy())
	controllerutil.RemoveFinalizer(tool, eirenyx.ToolFinalizer)
	if err := r.Patch(ctx, tool, patch); err != nil {
		return CompleteWithError(err)
	}

	log.Info("Finalizer removed, deletion completed", "tool", tool.Name)
	return Complete()
}

func (r *ToolReconciler) setFailedCondition(ctx context.Context, tool *eirenyx.Tool, reason, message string) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Error(nil, message, "tool", tool.Name)

	meta.SetStatusCondition(&tool.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})

	if err := r.Status().Update(ctx, tool); err != nil {
		log.Error(err, "Failed to update status condition")
		return RequeueAfter(5 * time.Second)
	}

	return Complete()
}

func (r *ToolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eirenyx.Tool{}).
		Named("tool").
		Complete(r)
}
