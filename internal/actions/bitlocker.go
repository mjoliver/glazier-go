package actions

import (
	"context"
	"fmt"

	"github.com/google/deck"
	"github.com/google/glazier/go/bitlocker"
	"gopkg.in/yaml.v3"
)

type BitLockerEnableConfig struct {
	Mode   string `yaml:"mode"`   // e.g., "tpm", "password"
	Backup bool   `yaml:"backup"` // Backup to AD?
}

type BitLockerEnable struct {
	Config BitLockerEnableConfig
}

func NewBitLockerEnable(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg BitLockerEnableConfig
	data, err := yaml.Marshal(yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &BitLockerEnable{Config: cfg}, nil
}

func (a *BitLockerEnable) Run(ctx context.Context) error {
	deck.Infof("Enabling BitLocker with mode: %s, backup: %v", a.Config.Mode, a.Config.Backup)

	// Example call to existing library
	if a.Config.Backup {
		if err := bitlocker.BackupToAD(); err != nil {
			return fmt.Errorf("failed to backup to AD: %w", err)
		}
	}

	// Note: Actual encryption logic would call other bitlocker package methods here
	// For now, mirroring the Python logic which calls similar helpers
	return nil
}

func (a *BitLockerEnable) Validate() error {
	if a.Config.Mode == "" {
		return fmt.Errorf("mode is required")
	}
	return nil
}

func init() {
	Register("bitlocker.enable", NewBitLockerEnable)
}
