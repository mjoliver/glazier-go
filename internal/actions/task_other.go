//go:build !windows

package actions

import (
	"context"
	"fmt"
)

func (a *Task) Run(ctx context.Context) error {
	return fmt.Errorf("task creation is only supported on Windows")
}
