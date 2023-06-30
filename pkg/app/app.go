package app

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chart"
	"kcl-lang.io/helm-kcl/pkg/config"
	"kcl-lang.io/helm-kcl/pkg/helm"
	"kcl-lang.io/krm-kcl/pkg/process"
)

// App is the main application object.
type App struct {
	helmBinary string
	logger     *zap.SugaredLogger
	render     helm.Render
}

// Template of App run the
func (app *App) Template(templateImpl *config.TemplateImpl) error {
	kclRun, err := config.FromFile(templateImpl.File)
	if err != nil {
		return err
	}
	for _, repo := range kclRun.Repositories {
		path, err := app.chartPathFromRepo(templateImpl.File, repo)
		if err != nil {
			return err
		}
		if err := app.template(templateImpl.File, repo.Name, path); err != nil {
			return err
		}
	}
	return nil
}

func (app *App) chartPathFromRepo(file string, repo config.RepositorySpec) (path string, err error) {
	if repo.URL != "" {
		path = repo.URL
	} else if repo.Path != "" {
		path = repo.Path
		if !filepath.IsAbs(repo.Path) {
			path = filepath.Join(filepath.Dir(file), repo.Path)
		}
	} else {
		return "", errors.New("no valid helm chart path, it should be from a local path or a url")
	}
	return path, nil
}

func (app *App) template(kclRunFile, release, chartPath string) error {
	// Generate Kubernetes manifests from helm charts.
	manifests, err := app.renderManifests(release, chartPath)
	if err != nil {
		return err
	}
	// KCL function config
	fnCfg, err := os.ReadFile(kclRunFile)
	if err != nil {
		return err
	}
	result, err := app.doMutate(manifests, fnCfg)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

// Generate Kubernetes manifests from helm charts.
func (app *App) renderManifests(release, chartPath string) ([]byte, error) {
	var chart *chart.Chart
	_, err := url.Parse(chartPath)
	// Load from url
	if err != nil {
		// Load from url
		chart, err = app.render.LoadChartFromRemoteCharts(chartPath)
		if err != nil {
			return nil, err
		}
	} else {
		// Load from local path
		chart, err = app.render.LoadChartFromLocalDirectory(chartPath)
		if err != nil {
			return nil, err
		}
	}
	manifests, err := app.render.GenerateManifests(release, fn.DefaultNamespace, chart, nil)
	if err != nil {
		return nil, err
	}
	return manifests, nil
}

func (app *App) doMutate(manifests, fnCfg []byte) (string, error) {
	items, err := fn.ParseKubeObjects(manifests)
	if err != nil {
		return "", err
	}
	functionConfig, err := fn.ParseKubeObject(fnCfg)
	if err != nil {
		return "", err
	}
	// Construct resource list.
	resourceList := &fn.ResourceList{
		Items:          items,
		FunctionConfig: functionConfig,
	}
	result, err := process.Process(resourceList)
	if err != nil {
		return "", err
	}
	if !result {
		return "", errors.New(resourceList.Results.Error())
	}
	return resourceList.Items.String(), nil
}
