//go:build !windows

package actions

import (
	"context"
	"fmt"
)

func (a *DomainJoin) Run(ctx context.Context) error {
	return fmt.Errorf("domain join is only supported on Windows")
}
