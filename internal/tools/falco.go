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
	falcoNamespace   = "falco"
	falcoReleaseName = "falco"
	falcoRepoName    = "falcosecurity"
	falcoRepoURL     = "https://falcosecurity.github.io/charts"
	falcoChart       = "falcosecurity/falco"
	falcoDaemonSet   = "falco"
)

type FalcoService struct {
	K8sClient *k8s.Client
}

func (f *FalcoService) Name() string {
	return "falco"
}

func (f *FalcoService) EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Installing or upgrading Falco")

	ns := resolvedNamespace(tool.Spec.Namespace, falcoNamespace)

	if err := f.K8sClient.EnsureNamespace(ctx, ns); err != nil {
		return fmt.Errorf("ensuring falco namespace: %w", err)
	}

	manager := helm.NewHelmAdminManager(ns, falcoRepoName, falcoRepoURL, falcoChart, falcoReleaseName)
	if err := manager.InstallOrUpgrade(); err != nil {
		return fmt.Errorf("installing falco: %w", err)
	}

	return nil
}

func (f *FalcoService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Uninstalling Falco")

	managedNamespace := tool.Spec.Namespace == ""
	ns := resolvedNamespace(tool.Spec.Namespace, falcoNamespace)

	manager := helm.NewHelmDeleteManager(ns, falcoReleaseName)
	if err := manager.Uninstall(); err != nil && !helm.IsReleaseNotFound(err) {
		return fmt.Errorf("uninstalling falco: %w", err)
	}

	if managedNamespace {
		if err := f.K8sClient.DeleteNamespace(ctx, ns); err != nil {
			return fmt.Errorf("deleting falco namespace: %w", err)
		}
	}

	return nil
}

func (f *FalcoService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) (bool, error) {
	log := logf.FromContext(ctx)
	log.Info("Checking Falco health")

	ns := resolvedNamespace(tool.Spec.Namespace, falcoNamespace)

	ready, err := f.K8sClient.IsDaemonSetReady(ctx, ns, falcoDaemonSet)
	if err != nil {
		return false, fmt.Errorf("checking falco daemonset health: %w", err)
	}

	log.Info("Falco health check complete", "healthy", ready)
	return ready, nil
}
