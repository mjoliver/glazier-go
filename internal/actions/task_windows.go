//go:build windows

package actions

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/capnspacehook/taskmaster"
	"github.com/google/glazier/go/tasks"
)

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
	args := strings.Join(a.Config.Args, " ")

	// Call library
	return tasks.Create("id", a.Config.Command, args, a.Config.Name, "SYSTEM", trigger)
}
