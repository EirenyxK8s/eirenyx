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
	litmusReleaseName = "chaos"
	litmusChart       = "litmuschaos/litmus"
	litmusRepoName    = "litmuschaos"
	litmusRepoURL     = "https://litmuschaos.github.io/litmus-helm/"
	litmusDeployName  = "chaos-litmus-server"
)

type LitmusService struct {
	K8sClient *k8s.Client
}

func (l *LitmusService) Name() string { return "litmus" }

func (l *LitmusService) EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Installing or upgrading Litmus ChaosCenter")

	ns := resolvedNamespace(tool.Spec.Namespace, litmusNamespace)

	if err := l.K8sClient.EnsureNamespace(ctx, ns); err != nil {
		return fmt.Errorf("ensuring litmus namespace: %w", err)
	}

	manager := helm.NewHelmAdminManager(ns, litmusRepoName, litmusRepoURL, litmusChart, litmusReleaseName)
	manager.SetValues(litmusDefaultValues(litmusReleaseName, ns))

	if raw := tool.Spec.Values.Raw; len(raw) != 0 {
		var userValues map[string]interface{}
		if err := json.Unmarshal(raw, &userValues); err != nil {
			return fmt.Errorf("decoding helm values: %w", err)
		}
		if userValues == nil {
			return fmt.Errorf("helm values must be a JSON object, got null or non-object")
		}
		manager.MergeValues(userValues)
	}

	if err := manager.InstallOrUpgrade(); err != nil {
		return fmt.Errorf("installing litmus: %w", err)
	}

	return nil
}

func (l *LitmusService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Uninstalling Litmus ChaosCenter")

	managedNamespace := tool.Spec.Namespace == ""
	ns := resolvedNamespace(tool.Spec.Namespace, litmusNamespace)

	manager := helm.NewHelmDeleteManager(ns, litmusReleaseName)
	if err := manager.Uninstall(); err != nil && !helm.IsReleaseNotFound(err) {
		return fmt.Errorf("uninstalling litmus: %w", err)
	}

	if managedNamespace {
		if err := l.K8sClient.DeleteNamespace(ctx, ns); err != nil {
			return fmt.Errorf("deleting litmus namespace: %w", err)
		}
	}

	return nil
}

func (l *LitmusService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) (bool, error) {
	log := logf.FromContext(ctx)
	log.Info("Checking Litmus ChaosCenter health")

	ns := resolvedNamespace(tool.Spec.Namespace, litmusNamespace)

	healthy, err := l.K8sClient.IsDeploymentReady(ctx, ns, litmusDeployName)
	if err != nil {
		return false, fmt.Errorf("checking litmus deployment health: %w", err)
	}

	log.Info("Litmus ChaosCenter health check complete", "healthy", healthy)
	return healthy, nil
}

func litmusDefaultValues(releaseName, ns string) map[string]interface{} {
	return map[string]interface{}{
		"portal": map[string]interface{}{
			"frontend": map[string]interface{}{
				"service": map[string]interface{}{"type": "NodePort"},
			},
			"server": map[string]interface{}{
				"graphqlServer": map[string]interface{}{
					"genericEnv": map[string]interface{}{
						"CHAOS_CENTER_UI_ENDPOINT": fmt.Sprintf(
							"http://%s-litmus-frontend-service.%s.svc.cluster.local:9091",
							releaseName, ns,
						),
					},
				},
			},
		},
	}
}
