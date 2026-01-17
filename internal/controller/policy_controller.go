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

	policyfactory "github.com/EirenyxK8s/eirenyx/internal/policy"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	eirenyxv1alpha1 "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
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
	_ = logf.FromContext(ctx)

	var policy eirenyxv1alpha1.Policy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	engine, err := policyfactory.NewEngine(&policy, policyfactory.Dependencies{
		Client: r.Client,
		Scheme: r.Scheme,
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	if !policy.Spec.Base.Enabled {
		return ctrl.Result{}, engine.Cleanup(ctx, &policy)
	}

	if err := engine.Validate(&policy); err != nil {
		return ctrl.Result{}, err
	}

	if err := engine.Reconcile(ctx, &policy); err != nil {
		return ctrl.Result{}, err
	}

	reportName, err := engine.GenerateReport(ctx, &policy)
	if err == nil {
		policy.Status.Base.LastReport = reportName
		_ = r.Status().Update(ctx, &policy)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eirenyxv1alpha1.Policy{}).
		Named("policy").
		Complete(r)
}
