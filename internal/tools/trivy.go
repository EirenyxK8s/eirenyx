package tools

import (
	"context"
	"fmt"
	"time"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/client/k8s"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmloader "helm.sh/helm/v3/pkg/chart/loader"
	helmcli "helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	namespace   = "trivy-system"
	releaseName = "trivy-operator"
)

type TrivyService struct {
}

func (t *TrivyService) Name() string {
	return "trivy"
}

func (t *TrivyService) EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error {
	if tool.Spec.InstallMethod != eirenyx.HelmInstall {
		return fmt.Errorf("unsupported install method: %s", tool.Spec.InstallMethod)
	}

	ns := tool.Spec.Namespace
	if ns == "" {
		ns = namespace
	}

	if err := k8s.EnsureK8sNamespace(ctx, ns); err != nil {
		return err
	}

	helmSpec := tool.Spec.Helm
	if helmSpec == nil {
		return fmt.Errorf("helm spec is required for trivy")
	}

	settings := helmcli.New()
	settings.SetNamespace(ns)

	actionConfig := new(helmaction.Configuration)
	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		ns,
		"secret",
		func(string, ...interface{}) {},
	); err != nil {
		return err
	}

	get := helmaction.NewGet(actionConfig)
	if _, err := get.Run(releaseName); err == nil {
		return nil
	}

	chartPath, err := helmaction.NewInstall(actionConfig).
		LocateChart(fmt.Sprintf("%s/%s", helmSpec.Repo, helmSpec.Chart), settings)
	if err != nil {
		return err
	}

	chart, err := helmloader.Load(chartPath)
	if err != nil {
		return err
	}

	values := map[string]interface{}{}
	if helmSpec.Values != nil {
		if err := yaml.Unmarshal(helmSpec.Values.Raw, &values); err != nil {
			return err
		}
	}

	install := helmaction.NewInstall(actionConfig)
	install.ReleaseName = releaseName
	install.Namespace = ns
	install.Wait = true
	install.Timeout = 5 * time.Minute

	if _, err := install.Run(chart, values); err != nil {
		return err
	}

	if done, _ := k8s.EnsureDeploymentRun(ctx, ns, litmusDeployName); !done {
		return fmt.Errorf("litmus server deployment not ready")
	}

	return nil
}

func (t *TrivyService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	if tool.Spec.InstallMethod != eirenyx.HelmInstall {
		return fmt.Errorf("unsupported install method: %s", tool.Spec.InstallMethod)
	}

	ns := tool.Spec.Namespace
	if ns == "" {
		ns = namespace
	}

	settings := helmcli.New()
	settings.SetNamespace(ns)

	actionConfig := new(helmaction.Configuration)
	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		ns,
		"secret",
		func(string, ...interface{}) {},
	); err != nil {
		return err
	}

	uninstall := helmaction.NewUninstall(actionConfig)
	uninstall.Timeout = 5 * time.Minute

	_, err := uninstall.Run(releaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall trivy-operator: %w", err)
	}

	return nil
}

func (t *TrivyService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) bool {
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = namespace
	}

	done, _ := k8s.EnsureDeploymentRun(ctx, ns, litmusDeployName)
	return done
}
