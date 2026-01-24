/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	policyfactory "github.com/EirenyxK8s/eirenyx/internal/policy"
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

// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=policies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=policies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=policies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
func (r *PolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Starting Policy Reconciliation")

	var policy eirenyx.Policy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		return CompleteWithError(client.IgnoreNotFound(err))
	}

	var tool eirenyx.Tool
	if err := r.Get(ctx, client.ObjectKey{
		Name:      string(policy.Spec.Base.Type),
		Namespace: policy.Namespace,
	}, &tool); err != nil {
		return CompleteWithError(err)
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
				return RequeueWithError(time.Second*5, err)
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

	if !policy.Spec.Base.Enabled {
		if err := engine.Cleanup(ctx, &policy); err != nil {
			return RequeueWithError(time.Second*5, err)
		}
		return Complete()
	}

	if err := engine.Validate(&policy); err != nil {
		return CompleteWithError(err)
	}

	if err := engine.Reconcile(ctx, &policy); err != nil {
		return RequeueWithError(time.Second*5, err)
	}

	reportName, err := engine.GenerateReport(ctx, &policy)
	if err == nil {
		policy.Status.Base.LastReport = reportName
		policy.Status.Base.ObservedGen = policy.Generation

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
