package controller

import (
	"context"
	"fmt"
	"time"

	policyfactory "github.com/EirenyxK8s/eirenyx/internal/policy"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
)

// PolicyReconciler reconciles a Policy object
type PolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *PolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Starting Policy Reconciliation")

	var policy eirenyx.Policy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		return CompleteWithError(client.IgnoreNotFound(err))
	}

	log.Info("Reconciling Policy With Spec: ", "policySpec", policy.Spec)

	var tool eirenyx.Tool
	if err := r.Get(ctx, client.ObjectKey{
		Name:      string(policy.Spec.Type),
		Namespace: policy.Namespace,
	}, &tool); err != nil {
		message := fmt.Sprintf("failed to get tool %q form namespace %q.", policy.Spec.Type, policy.Namespace)
		return CompleteWithError(errors.Wrap(err, message))
	}

	engine, err := policyfactory.NewEngine(&policy, policyfactory.Dependencies{
		Client: r.Client,
		Scheme: r.Scheme,
	})
	if err != nil {
		return CompleteWithError(err)
	}

	if !policy.DeletionTimestamp.IsZero() {
		log.Info("Policy is being deleted", "policy", policy.Name)

		if controllerutil.ContainsFinalizer(&policy, eirenyx.PolicyFinalizer) {
			if err := engine.Cleanup(ctx, &policy); err != nil {
				log.Error(err, "Failed to cleanup policy")
				return Requeue(time.Second * 5)
			}

			controllerutil.RemoveFinalizer(&policy, eirenyx.PolicyFinalizer)
			if err := r.Update(ctx, &policy); err != nil {
				return CompleteWithError(err)
			}
		}

		log.Info("Finalizer removed, deletion completed", "policy", policy.Name)
		return Complete()
	}

	if !controllerutil.ContainsFinalizer(&policy, eirenyx.PolicyFinalizer) {
		controllerutil.AddFinalizer(&policy, eirenyx.PolicyFinalizer)
		if err := r.Update(ctx, &policy); err != nil {
			return CompleteWithError(err)
		}
		return Complete()
	}

	has, err := controllerutil.HasOwnerReference(policy.OwnerReferences, &tool, r.Scheme)
	if err != nil {
		return CompleteWithError(err)
	}

	if !has {
		if err := controllerutil.SetOwnerReference(&tool, &policy, r.Scheme); err != nil {
			return CompleteWithError(err)
		}
		if err := r.Update(ctx, &policy); err != nil {
			return CompleteWithError(err)
		}
		return Complete()
	}

	if !policy.Spec.Enabled {
		if err := engine.Cleanup(ctx, &policy); err != nil {
			log.Error(err, "Failed to cleanup policy")
			return Requeue(time.Second * 5)
		}
		return Complete()
	}

	if err := engine.Validate(&policy); err != nil {
		return CompleteWithError(err)
	}

	if err := engine.Reconcile(ctx, &policy); err != nil {
		log.Error(err, "Failed to reconcile policy")
		return Requeue(time.Second * 5)
	}

	reportName, err := engine.GenerateReport(ctx, &policy)
	if err == nil {
		policy.Status.LastReport = reportName
		policy.Status.ObservedGen = policy.Generation

		if err := r.Status().Update(ctx, &policy); err != nil {
			return CompleteWithError(err)
		}
	}

	log.Info("Finished Policy Reconciliation", "policy", policy.Name)
	return Complete()
}

// SetupWithManager sets up the controller with the Manager.
func (r *PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eirenyx.Policy{}).
		Named("policy").
		Complete(r)
}
