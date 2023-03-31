package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"go.uber.org/zap"
	"kusionstack.io/helm-kcl/pkg/config"
	"kusionstack.io/helm-kcl/pkg/helm"
	"kusionstack.io/kpt-kcl-sdk/pkg/process"
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
		path := repo.Path
		if !filepath.IsAbs(repo.Path) {
			path = filepath.Join(filepath.Dir(templateImpl.File), repo.Path)
		}
		if err := app.template(templateImpl.File, repo.Name, path); err != nil {
			return err
		}
	}
	return nil
}

func (app *App) template(kclRunFile, release, chartDir string) error {
	// Generate Kubernetes manifests from helm charts.
	manifests, err := app.renderManifests(release, chartDir)
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
func (app *App) renderManifests(release, chartDir string) ([]byte, error) {
	chart, err := app.render.LoadChartFromLocalDirectory(chartDir)
	if err != nil {
		return nil, err
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
