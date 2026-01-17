package tools

import (
	"context"
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/client/helm"
	"github.com/EirenyxK8s/eirenyx/internal/client/k8s"
	"k8s.io/apimachinery/pkg/util/json"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	litmusNamespace   = "litmus"
	litmusReleaseName = "litmus"
	litmusDeployName  = "litmus-server"
)

type LitmusService struct {
}

func (l *LitmusService) Name() string {
	return "litmus"
}

func (l *LitmusService) EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Installing or upgrading Litmus Operator")
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = litmusNamespace
	}

	if err := k8s.EnsureK8sNamespace(ctx, ns); err != nil {
		return err
	}

	manager := helm.NewHelmAdminManager(
		ns,
		"litmuschaos",
		"https://litmuschaos.github.io/litmus-helm/",
		"litmuschaos/litmus",
		litmusReleaseName,
	)

	rawValues := tool.Spec.Values.Raw
	if len(rawValues) != 0 {
		var values map[string]interface{}
		if err := json.Unmarshal(rawValues, &values); err != nil {
			return fmt.Errorf("failed to decode helm values: %w", err)
		}
		manager.SetValues(values)
	}

	return manager.InstallOrUpgrade()
}

func (l *LitmusService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Uninstalling Litmus Operator")
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = litmusNamespace
	}

	manager := helm.NewHelmDeleteManager(ns, litmusReleaseName)
	return manager.Uninstall()
}

func (l *LitmusService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) bool {
	log := logf.FromContext(ctx)
	log.Info("Checking Litmus Operator health")
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = litmusNamespace
	}

	return k8s.EnsureDeploymentRun(ctx, ns, litmusDeployName)
}
