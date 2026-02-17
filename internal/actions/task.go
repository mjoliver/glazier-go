package actions

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
)

type TaskConfig struct {
	Name    string   `yaml:"name"`
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	Trigger string   `yaml:"trigger"` // "boot" or "time"
	Time    string   `yaml:"time"`    // for "time" trigger
}

func NewTask(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg TaskConfig
	data, err := yaml.Marshal(yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &Task{Config: cfg}, nil
}

type Task struct {
	Config TaskConfig
}

func (a *Task) Validate() error {
	if a.Config.Name == "" || a.Config.Command == "" {
		return fmt.Errorf("name and command are required")
	}
	return nil
}

func init() {
	Register("task.create", NewTask)
}
