package actions

import (
	"context"
	"fmt"
	"log"

	"github.com/google/glazier/go/power"
	"gopkg.in/yaml.v3"
)

type PowerConfig struct {
	Type   string `yaml:"type"`   // "reboot" or "shutdown"
	Delay  int    `yaml:"delay"`  // seconds
	Reason string `yaml:"reason"` // e.g. "maintenance", "installation"
	Force  bool   `yaml:"force"`
}

func NewPower(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg PowerConfig
	if str, ok := yamlData.(string); ok {
		cfg.Type = str
	} else {
		data, err := yaml.Marshal(yamlData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	return &Power{Config: cfg}, nil
}

type Power struct {
	Config PowerConfig
}

func (a *Power) Run(ctx context.Context) error {
	log.Printf("Executing power action: %s", a.Config.Type)

	reason := power.SHTDN_REASON_MAJOR_NONE
	switch a.Config.Reason {
	case "maintenance":
		reason = power.SHTDN_REASON_MINOR_MAINTENANCE
	case "installation":
		reason = power.SHTDN_REASON_MINOR_INSTALLATION
	case "upgrade":
		reason = power.SHTDN_REASON_MINOR_UPGRADE
	}

	if a.Config.Type == "reboot" {
		return power.Reboot(reason, a.Config.Force)
	} else if a.Config.Type == "shutdown" {
		return power.Shutdown(reason, a.Config.Force)
	}

	return fmt.Errorf("unknown power type: %s", a.Config.Type)
}

func (a *Power) Validate() error {
	if a.Config.Type != "reboot" && a.Config.Type != "shutdown" {
		return fmt.Errorf("invalid power type: %s", a.Config.Type)
	}
	return nil
}

func init() {
	Register("system.power", NewPower)
}
