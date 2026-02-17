package actions

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/capnspacehook/taskmaster"
	"github.com/google/glazier/go/tasks"
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

func (a *Task) Run(ctx context.Context) error {
	log.Printf("Creating scheduled task: %s", a.Config.Name)

	// Determine trigger
	var trigger taskmaster.Trigger
	if a.Config.Trigger == "boot" {
		trigger = taskmaster.BootTrigger{
			TaskTrigger: taskmaster.TaskTrigger{Enabled: true, ID: "boot"},
		}
	} else {
		// Default or time
		start := time.Now().Add(1 * time.Minute)
		trigger = taskmaster.TimeTrigger{
			TaskTrigger: taskmaster.TaskTrigger{
				Enabled:       true,
				ID:            "time",
				StartBoundary: start,
			},
		}
	}

	// Join args
	args := ""
	for _, arg := range a.Config.Args {
		args += " " + arg
	}

	// Call library
	return tasks.Create("id", a.Config.Command, args, a.Config.Name, "SYSTEM", trigger)
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
