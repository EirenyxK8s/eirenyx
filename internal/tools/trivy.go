package tools

import (
	"context"
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/client/helm"
	"github.com/EirenyxK8s/eirenyx/internal/client/k8s"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	trivyNamespace   = "trivy-system"
	trivyReleaseName = "trivy-operator"
	trivyRepoName    = "aqua"
	trivyRepoURL     = "https://aquasecurity.github.io/helm-charts/"
	trivyChart       = "aqua/trivy-operator"
)

type TrivyService struct {
	K8sClient *k8s.Client
}

func (t *TrivyService) Name() string { return "trivy" }

func (t *TrivyService) EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Installing or upgrading Trivy Operator")

	ns := resolvedNamespace(tool.Spec.Namespace, trivyNamespace)

	if err := t.K8sClient.EnsureNamespace(ctx, ns); err != nil {
		return fmt.Errorf("ensuring trivy namespace: %w", err)
	}

	manager := helm.NewHelmAdminManager(ns, trivyRepoName, trivyRepoURL, trivyChart, trivyReleaseName)
	if err := manager.InstallOrUpgrade(); err != nil {
		return fmt.Errorf("installing trivy operator: %w", err)
	}

	return nil
}

func (t *TrivyService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Uninstalling Trivy Operator")

	managedNamespace := tool.Spec.Namespace == ""
	ns := resolvedNamespace(tool.Spec.Namespace, trivyNamespace)

	manager := helm.NewHelmDeleteManager(ns, trivyReleaseName)
	if err := manager.Uninstall(); err != nil && !helm.IsReleaseNotFound(err) {
		return fmt.Errorf("uninstalling trivy operator: %w", err)
	}

	if managedNamespace {
		if err := t.K8sClient.DeleteNamespace(ctx, ns); err != nil {
			return fmt.Errorf("deleting trivy namespace: %w", err)
		}
	}

	return nil
}

func (t *TrivyService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) (bool, error) {
	log := logf.FromContext(ctx)
	log.Info("Checking Trivy Operator health")

	ns := resolvedNamespace(tool.Spec.Namespace, trivyNamespace)

	healthy, err := t.K8sClient.IsDeploymentReady(ctx, ns, trivyReleaseName)
	if err != nil {
		return false, fmt.Errorf("checking trivy deployment health: %w", err)
	}

	log.Info("Trivy Operator health check complete", "healthy", healthy)
	return healthy, nil
}
