package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"go.uber.org/zap"
	"k8s.io/helm/pkg/helm"
	"kusionstack.io/helm-kcl/pkg/config"
	"kusionstack.io/kpt-kcl-sdk/pkg/process"
)

// App is the main application object.
type App struct {
	helmBinary string
	logger     *zap.SugaredLogger
	helm       helm.Interface
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
		if IsHelm3() {
			if err := app.runHelm3Template(templateImpl.File, repo.Name, path); err != nil {
				return err
			}
		} else {
			return errors.New("Helm 2 is not supported yet")
		}
	}
	return nil
}

func (app *App) runHelm3Template(kclRunFile, release, chart string) error {
	if err := compatibleHelm3Version(); err != nil {
		app.logger.Error(err)
		return err
	}
	e := executor{
		release: release,
		chart:   chart,
		dryRun:  true,
	}
	// Kubernetes manifests
	manifests, err := e.template(false)
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
