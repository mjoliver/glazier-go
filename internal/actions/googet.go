package actions

import (
	"context"
	"fmt"
	"log"

	"github.com/google/glazier/go/googet"
	"gopkg.in/yaml.v3"
)

type GooGetInstallConfig struct {
	Packages  []string `yaml:"packages"`
	Reinstall bool     `yaml:"reinstall"`
	DBOnly    bool     `yaml:"db_only"` // Update local DB only?
}

type GooGetInstall struct {
	Config GooGetInstallConfig
}

func NewGooGetInstall(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg GooGetInstallConfig
	// If the YAML data is just a list of strings, handle that shorthand
	if list, ok := yamlData.([]interface{}); ok {
		for _, item := range list {
			if str, ok := item.(string); ok {
				cfg.Packages = append(cfg.Packages, str)
			}
		}
	} else {
		// otherwise treat as structured config
		data, err := yaml.Marshal(yamlData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	return &GooGetInstall{Config: cfg}, nil
}

func (a *GooGetInstall) Run(ctx context.Context) error {
	log.Printf("Installing GooGet packages: %v", a.Config.Packages)
	
	for _, pkg := range a.Config.Packages {
		// Call existing library
		// Note: existing library accepts package, sources, reinstall, dbOnly, config
		if err := googet.Install(pkg, "", a.Config.Reinstall, a.Config.DBOnly, nil); err != nil {
			return fmt.Errorf("failed to install package %s: %w", pkg, err)
		}
	}
	return nil
}

func (a *GooGetInstall) Validate() error {
	if len(a.Config.Packages) == 0 {
		return fmt.Errorf("packages list is empty")
	}
	return nil
}

func init() {
	Register("googet.install", NewGooGetInstall)
}
