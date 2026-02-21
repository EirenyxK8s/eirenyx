package tools

import (
	"context"
	"fmt"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/client/helm"
	"github.com/EirenyxK8s/eirenyx/internal/client/k8s"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/json"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	litmusNamespace   = "litmus"
	litmusReleaseName = "chaos" // MUST match working CLI
	litmusChart       = "litmuschaos/litmus"
	litmusRepoName    = "litmuschaos"
	litmusRepoURL     = "https://litmuschaos.github.io/litmus-helm/"
	litmusDeployName  = "litmus-server"
)

type LitmusService struct{}

func (l *LitmusService) Name() string {
	return "litmus"
}

func (l *LitmusService) EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Installing or upgrading Litmus ChaosCenter")

	ns := tool.Spec.Namespace
	if ns == "" {
		ns = litmusNamespace
	}

	if err := k8s.EnsureK8sNamespace(ctx, ns); err != nil {
		return errors.Wrap(err, "failed to ensure litmus namespace")
	}

	manager := helm.NewHelmAdminManager(
		ns,
		litmusRepoName,
		litmusRepoURL,
		litmusChart,
		litmusReleaseName,
	)

	defaultValues := map[string]interface{}{
		"portal": map[string]interface{}{
			"frontend": map[string]interface{}{
				"service": map[string]interface{}{
					"type": "NodePort",
				},
			},
			"server": map[string]interface{}{
				"graphqlServer": map[string]interface{}{
					"genericEnv": map[string]interface{}{
						"CHAOS_CENTER_UI_ENDPOINT": fmt.Sprintf(
							"http://%s-litmus-frontend-service.%s.svc.cluster.local:9091",
							litmusReleaseName,
							ns,
						),
					},
				},
			},
		},
	}

	manager.SetValues(defaultValues)
	if raw := tool.Spec.Values.Raw; len(raw) != 0 {
		var userValues map[string]interface{}
		if err := json.Unmarshal(raw, &userValues); err != nil {
			return errors.Wrap(err, "failed to decode helm values")
		}
		manager.MergeValues(userValues)
	}

	return manager.InstallOrUpgrade()
}

func (l *LitmusService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Uninstalling Litmus ChaosCenter")

	ns := tool.Spec.Namespace
	if ns == "" {
		ns = litmusNamespace
	}

	manager := helm.NewHelmDeleteManager(ns, litmusReleaseName)

	if err := manager.Uninstall(); err != nil {
		return errors.Wrap(err, "failed to uninstall Litmus via Helm")
	}

	return k8s.EnsureNamespaceDeleted(ctx, ns)
}

func (l *LitmusService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) bool {
	log := logf.FromContext(ctx)
	log.Info("Checking Litmus ChaosCenter health")

	ns := tool.Spec.Namespace
	if ns == "" {
		ns = litmusNamespace
	}

	return k8s.EnsureDeploymentRun(ctx, ns, litmusDeployName)
}
