package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"go.uber.org/zap"
	"k8s.io/helm/pkg/helm"
	"kusionstack.io/helm-kcl/pkg/config"
	"kusionstack.io/kpt-kcl-sdk/pkg/process"
)

var CleanWaitGroup sync.WaitGroup

// App is the main application object.
type App struct {
	helmBinary string
	logger     *zap.SugaredLogger
	helm       helm.Interface
}

type HelmRelease struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Enabled   bool   `json:"enabled"`
	Installed bool   `json:"installed"`
	Labels    string `json:"labels"`
	Chart     string `json:"chart"`
	Version   string `json:"version"`
}

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
	d := diffCmd{
		release: release,
		chart:   chart,
		dryRun:  true,
	}
	// Kubernetes manifests
	template, err := d.template(false)
	if err != nil {
		return err
	}
	// KCL function config
	fnCfgBytes, err := os.ReadFile(kclRunFile)
	if err != nil {
		return err
	}
	items, err := fn.ParseKubeObjects(template)
	if err != nil {
		return err
	}
	fnCfg, err := fn.ParseKubeObject(fnCfgBytes)
	if err != nil {
		return err
	}
	resourceList := &fn.ResourceList{
		Items:          items,
		FunctionConfig: fnCfg,
	}
	result, err := process.Process(resourceList)
	if err != nil {
		return err
	}
	if !result {
		return errors.New(resourceList.Results.Error())
	}
	fmt.Println(resourceList.Items.String())
	return nil
}
