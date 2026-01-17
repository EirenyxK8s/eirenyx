package tools

import (
	"context"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/client/helm"
	"github.com/EirenyxK8s/eirenyx/internal/client/k8s"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	trivyNamespace   = "trivy-system"
	trivyReleaseName = "trivy-operator"
)

type TrivyService struct {
}

func (t *TrivyService) Name() string {
	return "trivy"
}

func (t *TrivyService) EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Installing or upgrading Trivy Operator")
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = trivyNamespace
	}

	if err := k8s.EnsureK8sNamespace(ctx, ns); err != nil {
		return err
	}

	manager := helm.NewHelmAdminManager(
		ns,
		"aqua",
		"https://aquasecurity.github.io/helm-charts/",
		"aqua/trivy-operator",
		trivyReleaseName,
	)
	return manager.InstallOrUpgrade()
}

func (t *TrivyService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Uninstalling Trivy Operator")
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = trivyNamespace
	}

	manager := helm.NewHelmDeleteManager(ns, trivyReleaseName)
	return manager.Uninstall()
}

func (t *TrivyService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) bool {
	log := logf.FromContext(ctx)
	log.Info("Checking Trivy Operator health")
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = trivyNamespace
	}

	return k8s.EnsureDeploymentRun(ctx, ns, trivyReleaseName)
}
