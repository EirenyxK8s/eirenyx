package helm

import (
	"os"
	"strings"
	"time"

	helmaction "helm.sh/helm/v3/pkg/action"
	helmloader "helm.sh/helm/v3/pkg/chart/loader"
	helmcli "helm.sh/helm/v3/pkg/cli"
	helmgetter "helm.sh/helm/v3/pkg/getter"
	helmrepo "helm.sh/helm/v3/pkg/repo"
)

type Manager struct {
	Namespace   string
	RepoName    string
	RepoURL     string
	ChartRef    string
	ReleaseName string
	Values      map[string]interface{}
}

func NewHelmAdminManager(namespace, repoName, repoUrl, chartRef, releaseName string) *Manager {
	return &Manager{
		Namespace:   namespace,
		RepoName:    repoName,
		RepoURL:     repoUrl,
		ChartRef:    chartRef,
		ReleaseName: releaseName,
	}
}

func (i *Manager) SetValues(values map[string]interface{}) {
	i.Values = values
}

func NewHelmDeleteManager(namespace, releaseName string) *Manager {
	return &Manager{
		Namespace:   namespace,
		ReleaseName: releaseName,
	}
}

func (i *Manager) InstallOrUpgrade() error {
	settings, err := i.settings()
	if err != nil {
		return err
	}

	if err := i.ensureRepo(settings); err != nil {
		return err
	}

	cfg, err := i.actionConfig(settings)
	if err != nil {
		return err
	}

	get := helmaction.NewGet(cfg)
	if _, err := get.Run(i.ReleaseName); err != nil {
		install := helmaction.NewInstall(cfg)
		install.ReleaseName = i.ReleaseName
		install.Namespace = i.Namespace
		install.Wait = true
		install.Timeout = 5 * time.Minute

		chartPath, err := install.LocateChart(i.ChartRef, settings)
		if err != nil {
			return err
		}

		chart, err := helmloader.Load(chartPath)
		if err != nil {
			return err
		}

		_, err = install.Run(chart, i.Values)
		return err
	}

	upgrade := helmaction.NewUpgrade(cfg)
	upgrade.Namespace = i.Namespace
	upgrade.Wait = true
	upgrade.Timeout = 5 * time.Minute

	chartPath, err := upgrade.LocateChart(i.ChartRef, settings)
	if err != nil {
		return err
	}

	chart, err := helmloader.Load(chartPath)
	if err != nil {
		return err
	}

	_, err = upgrade.Run(i.ReleaseName, chart, i.Values)
	return err
}

func (i *Manager) Uninstall() error {
	settings, err := i.settings()
	if err != nil {
		return err
	}

	cfg, err := i.actionConfig(settings)
	if err != nil {
		return err
	}

	uninstall := helmaction.NewUninstall(cfg)
	uninstall.Timeout = 5 * time.Minute

	if _, err := uninstall.Run(i.ReleaseName); err != nil {
		if strings.Contains(err.Error(), "release: not found") ||
			strings.Contains(err.Error(), "Release not loaded") {
			return nil
		}
		return err
	}
	return err
}

func (i *Manager) MergeValues(user map[string]interface{}) {
	if i.Values == nil {
		i.Values = map[string]interface{}{}
	}
	i.Values = deepMerge(i.Values, user)
}

func (i *Manager) settings() (*helmcli.EnvSettings, error) {
	settings := helmcli.New()
	settings.SetNamespace(i.Namespace)
	settings.RepositoryConfig = "/tmp/helm/repositories.yaml"
	settings.RepositoryCache = "/tmp/helm/cache"

	if err := os.MkdirAll(settings.RepositoryCache, 0755); err != nil {
		return nil, err
	}

	if _, err := os.Stat(settings.RepositoryConfig); os.IsNotExist(err) {
		f, err := os.Create(settings.RepositoryConfig)
		if err != nil {
			return nil, err
		}
		_ = f.Close()
	}

	return settings, nil
}

func (i *Manager) actionConfig(settings *helmcli.EnvSettings) (*helmaction.Configuration, error) {
	cfg := new(helmaction.Configuration)
	if err := cfg.Init(
		settings.RESTClientGetter(),
		i.Namespace,
		"secret",
		func(string, ...interface{}) {},
	); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (i *Manager) ensureRepo(settings *helmcli.EnvSettings) error {
	repoFile, err := helmrepo.LoadFile(settings.RepositoryConfig)
	if err != nil {
		return err
	}

	if !repoFile.Has(i.RepoName) {
		repoFile.Update(&helmrepo.Entry{
			Name: i.RepoName,
			URL:  i.RepoURL,
		})
		if err := repoFile.WriteFile(settings.RepositoryConfig, 0644); err != nil {
			return err
		}
	}

	entry := repoFile.Get(i.RepoName)
	repo, err := helmrepo.NewChartRepository(entry, helmgetter.All(settings))
	if err != nil {
		return err
	}

	repo.CachePath = settings.RepositoryCache
	_, err = repo.DownloadIndexFile()
	return err
}

func deepMerge(dst, src map[string]interface{}) map[string]interface{} {
	for key, srcVal := range src {
		if dstVal, exists := dst[key]; exists {
			dstMap, dstIsMap := dstVal.(map[string]interface{})
			srcMap, srcIsMap := srcVal.(map[string]interface{})

			if dstIsMap && srcIsMap {
				dst[key] = deepMerge(dstMap, srcMap)
				continue
			}
			dst[key] = srcVal
		} else {
			dst[key] = srcVal
		}
	}
	return dst
}
