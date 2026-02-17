package actions

import (
	"context"
	"fmt"
	"log"

	"gopkg.in/yaml.v3"
)

// RegistryConfig holds config for registry actions.
// All operations target HKLM (HKEY_LOCAL_MACHINE).
type RegistryConfig struct {
	Path  string      `yaml:"path"`  // e.g. "SOFTWARE\\Glazier"
	Name  string      `yaml:"name"`  // value name
	Value interface{} `yaml:"value"` // for set: string, int, or []string
	Type  string      `yaml:"type"`  // "string", "dword", "multi_string", "binary"
}

// --- registry.set ---

type RegistrySet struct{ Config RegistryConfig }

func NewRegistrySet(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg RegistryConfig
	data, _ := yaml.Marshal(yamlData)
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("registry.set: %w", err)
	}
	return &RegistrySet{Config: cfg}, nil
}

func (a *RegistrySet) Validate() error {
	if a.Config.Path == "" || a.Config.Name == "" {
		return fmt.Errorf("registry.set: path and name are required")
	}
	if a.Config.Type == "" {
		a.Config.Type = "string"
	}
	return nil
}

// --- registry.delete ---

type RegistryDelete struct{ Config RegistryConfig }

func NewRegistryDelete(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg RegistryConfig
	data, _ := yaml.Marshal(yamlData)
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("registry.delete: %w", err)
	}
	return &RegistryDelete{Config: cfg}, nil
}

func (a *RegistryDelete) Validate() error {
	if a.Config.Path == "" || a.Config.Name == "" {
		return fmt.Errorf("registry.delete: path and name are required")
	}
	return nil
}

// --- registry.get ---

type RegistryGet struct {
	Config RegistryConfig
	Result string // populated after Run on Windows
}

func NewRegistryGet(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg RegistryConfig
	data, _ := yaml.Marshal(yamlData)
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("registry.get: %w", err)
	}
	return &RegistryGet{Config: cfg}, nil
}

func (a *RegistryGet) Validate() error {
	if a.Config.Path == "" || a.Config.Name == "" {
		return fmt.Errorf("registry.get: path and name are required")
	}
	return nil
}

func init() {
	Register("registry.set", NewRegistrySet)
	Register("registry.delete", NewRegistryDelete)
	Register("registry.get", NewRegistryGet)
	_ = log.Printf
}
