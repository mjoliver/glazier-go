//go:build !windows

package actions

import (
	"context"
	"fmt"
)

func (a *RegistrySet) Run(ctx context.Context) error {
	return fmt.Errorf("registry operations are only supported on Windows")
}

func (a *RegistryDelete) Run(ctx context.Context) error {
	return fmt.Errorf("registry operations are only supported on Windows")
}

func (a *RegistryGet) Run(ctx context.Context) error {
	return fmt.Errorf("registry operations are only supported on Windows")
}
