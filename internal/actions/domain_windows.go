//go:build windows

package actions

import (
	"context"

	"github.com/google/deck"
	join "github.com/google/glazier/go/domain"
)

func (a *DomainJoin) Run(ctx context.Context) error {
	deck.Infof("Joining domain: %s (OU: %s)", a.Config.Domain, a.Config.OU)
	return join.Domain(a.Config.Domain, a.Config.OU, a.Config.User, a.Config.Password, join.JoinDomainFlag)
}
