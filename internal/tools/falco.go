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
	falcoNamespace   = "falco"
	falcoReleaseName = "falco"
	falcoDaemonSet   = "falco"
)

type FalcoService struct {
}

func (f *FalcoService) Name() string {
	return "falco"
}

func (f *FalcoService) EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error {
	if tool.Spec.InstallMethod != eirenyx.HelmInstall {
		return fmt.Errorf("unsupported install method: %s", tool.Spec.InstallMethod)
	}

	ns := tool.Spec.Namespace
	if ns == "" {
		ns = falcoNamespace
	}

	if err := k8s.EnsureK8sNamespace(ctx, ns); err != nil {
		return err
	}

	helmSpec := tool.Spec.Helm
	if helmSpec == nil {
		return fmt.Errorf("helm spec is required for falco")
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

	// Idempotency check
	get := helmaction.NewGet(actionConfig)
	if _, err := get.Run(falcoReleaseName); err == nil {
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
	install.ReleaseName = falcoReleaseName
	install.Namespace = ns
	install.Wait = true
	install.Timeout = 10 * time.Minute

	if _, err := install.Run(chart, values); err != nil {
		return fmt.Errorf("failed to install falco: %w", err)
	}

	// Minimal readiness check: DaemonSet exists and has ready pods
	ok, err := k8s.IsDaemonSetReady(ctx, ns, falcoDaemonSet)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("falco daemonset not ready")
	}

	return nil
}

func (f *FalcoService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	if tool.Spec.InstallMethod != eirenyx.HelmInstall {
		return fmt.Errorf("unsupported install method: %s", tool.Spec.InstallMethod)
	}

	ns := tool.Spec.Namespace
	if ns == "" {
		ns = falcoNamespace
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

	_, err := uninstall.Run(falcoReleaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall falco: %w", err)
	}

	return nil
}

func (f *FalcoService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) bool {
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = falcoNamespace
	}

	ok, err := k8s.IsDaemonSetReady(ctx, ns, falcoDaemonSet)
	if err != nil {
		return false
	}

	return ok
}
