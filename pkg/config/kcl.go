package config

import (
	"os"

	"github.com/KusionStack/krm-kcl/pkg/config"
	"gopkg.in/yaml.v2"
)

// KCLRun is a custom resource to provider Helm kcl config including KCL source and params.
type KCLRun struct {
	config.KCLRun `json:",inline" yaml:",inline"`
	Repositories  []RepositorySpec `yaml:"repositories,omitempty"`
}

func FromFile(file string) (*KCLRun, error) {
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var config KCLRun
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
