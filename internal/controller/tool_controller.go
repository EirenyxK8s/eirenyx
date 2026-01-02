package controller

import (
	"context"
	"time"

	"github.com/EirenyxK8s/eirenyx/internal/tools"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

// ToolReconciler reconciles a Tool object
type ToolReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Service map[eirenyx.ToolType]tools.ToolService
}

// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=tools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=tools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=tools/finalizers,verbs=update

func (r *ToolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var tool eirenyx.Tool
	if err := r.Get(ctx, req.NamespacedName, &tool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	service, ok := r.Service[tool.Spec.Type]
	if !ok {
		log.Error(nil, "Tool type not found", "tool", tool.Spec.Type)
		return ctrl.Result{}, nil
	}

	if tool.Spec.Enabled {
		if err := service.EnsureInstalled(ctx, &tool); err != nil {
			log.Error(err, "Failed to install tool", "tool", tool.Spec.Type)
			return ctrl.Result{RequeueAfter: time.Microsecond}, err
		}
	} else {
		if err := service.EnsureUninstalled(ctx, &tool); err != nil {
			log.Error(err, "Failed to uninstall tool", "tool", tool.Spec.Type)
			return ctrl.Result{RequeueAfter: time.Microsecond}, err
		}
	}

	healthy, err := service.CheckHealth(ctx, &tool)
	if err != nil {
		log.Error(err, "Failed to check health of tool", "tool", tool.Spec.Type)
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	tool.Status.Installed = tool.Spec.Enabled
	tool.Status.Healthy = healthy
	_ = r.Status().Update(ctx, &tool)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ToolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eirenyx.Tool{}).
		Named("tool").
		Complete(r)
}
