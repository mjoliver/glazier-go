//go:build !windows

package actions

import (
	"context"
	"fmt"
)

func (a *Power) Run(ctx context.Context) error {
	return fmt.Errorf("power actions are only supported on Windows")
}
