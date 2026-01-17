package tools

import (
	"context"
	"fmt"
	"os"
	"time"

	eirenyx "github.com/EirenyxK8s/eirenyx/api/v1alpha1"
	"github.com/EirenyxK8s/eirenyx/internal/client/k8s"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmloader "helm.sh/helm/v3/pkg/chart/loader"
	helmcli "helm.sh/helm/v3/pkg/cli"
	helmgetter "helm.sh/helm/v3/pkg/getter"
	helmrepo "helm.sh/helm/v3/pkg/repo"
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

	settings := helmcli.New()
	settings.SetNamespace(ns)
	settings.RepositoryConfig = "/tmp/helm/repositories.yaml"
	settings.RepositoryCache = "/tmp/helm/cache"

	log.Info("Preparing helm repository and cache directories")
	if err := os.MkdirAll(settings.RepositoryCache, 0755); err != nil {
		return err
	}
	if _, err := os.Stat(settings.RepositoryConfig); os.IsNotExist(err) {
		f, err := os.Create(settings.RepositoryConfig)
		if err != nil {
			return err
		}
		_ = f.Close()
	}

	repoName := "aqua"
	repoURL := "https://aquasecurity.github.io/helm-charts/"

	log.Info("Preparing trivy helm repository")
	repoFile, err := helmrepo.LoadFile(settings.RepositoryConfig)
	if err != nil {
		return err
	}

	if !repoFile.Has(repoName) {
		repoFile.Update(&helmrepo.Entry{Name: repoName, URL: repoURL})
		if err := repoFile.WriteFile(settings.RepositoryConfig, 0644); err != nil {
			return err
		}
	}

	entry := repoFile.Get(repoName)
	repo, err := helmrepo.NewChartRepository(entry, helmgetter.All(settings))
	if err != nil {
		return err
	}
	repo.CachePath = settings.RepositoryCache

	if _, err := repo.DownloadIndexFile(); err != nil {
		return err
	}

	actionConfig := new(helmaction.Configuration)
	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		ns,
		"secret",
		func(string, ...interface{}) {},
	); err != nil {
		return err
	}

	chartRef := "aqua/trivy-operator"

	get := helmaction.NewGet(actionConfig)
	if _, err = get.Run(trivyReleaseName); err != nil {
		log.Info("Installing Trivy Operator")
		install := helmaction.NewInstall(actionConfig)
		install.ReleaseName = trivyReleaseName
		install.Namespace = ns
		install.Wait = true
		install.Timeout = 5 * time.Minute

		chartPath, err := install.LocateChart(chartRef, settings)
		if err != nil {
			return err
		}

		chart, err := helmloader.Load(chartPath)
		if err != nil {
			return err
		}

		if _, err := install.Run(chart, map[string]interface{}{}); err != nil {
			return err
		}
	} else {
		log.Info("Trivy Operator already installed, upgrading to latest version")
		upgrade := helmaction.NewUpgrade(actionConfig)
		upgrade.Namespace = ns
		upgrade.Wait = true
		upgrade.Timeout = 5 * time.Minute

		chartPath, err := upgrade.LocateChart(chartRef, settings)
		if err != nil {
			return err
		}

		chart, err := helmloader.Load(chartPath)
		if err != nil {
			return err
		}

		if _, err := upgrade.Run(trivyReleaseName, chart, map[string]interface{}{}); err != nil {
			return err
		}
	}
	
	log.Info("Trivy Operator installed successfully")
	return nil
}

func (t *TrivyService) EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error {
	log := logf.FromContext(ctx)
	log.Info("Uninstalling Trivy Operator")
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = trivyNamespace
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

	log.Info("Uninstalling trivy operator helm release")
	uninstall := helmaction.NewUninstall(actionConfig)
	uninstall.Timeout = 5 * time.Minute

	_, err := uninstall.Run(trivyReleaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall trivy-operator: %w", err)
	}

	log.Info("Trivy Operator uninstalled successfully")
	return nil
}

func (t *TrivyService) CheckHealth(ctx context.Context, tool *eirenyx.Tool) bool {
	log := logf.FromContext(ctx)
	log.Info("Checking Trivy Operator health")
	ns := tool.Spec.Namespace
	if ns == "" {
		ns = trivyNamespace
	}

	done, _ := k8s.EnsureDeploymentRun(ctx, ns, trivyReleaseName)
	return done
}
