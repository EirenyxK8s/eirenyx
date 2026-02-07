package tools

import (
	"context"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/client/helm"
	"github.com/EirenyxK8s/eirenyx/internal/client/k8s"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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
	log := logf.FromContext(ctx)
	log.Info("Installing or upgrading Falco")
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = falcoNamespace
	}

	if err := k8s.EnsureK8sNamespace(ctx, ns); err != nil {
		return err
	}

	manager := helm.NewHelmAdminManager(
		ns,
		"falcosecurity",
		"https://falcosecurity.github.io/charts",
		"falcosecurity/falco",
		falcoReleaseName,
	)
	return manager.InstallOrUpgrade()
}

func (f *FalcoService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Uninstalling Falco")
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = falcoNamespace
	}

	manager := helm.NewHelmDeleteManager(ns, falcoReleaseName)
	if err := manager.Uninstall(); err != nil {
		return errors.Wrap(err, "failed to uninstall Falco Operator via Helm")
	}

	if err := k8s.EnsureNamespaceDeleted(ctx, ns); err != nil {
		return err
	}
	return nil
}

func (f *FalcoService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) bool {
	log := logf.FromContext(ctx)
	log.Info("Checking Falco health")
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = falcoNamespace
	}

	return k8s.IsDaemonSetReady(ctx, ns, falcoDaemonSet)
}
