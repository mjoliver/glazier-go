//go:build windows

package actions

import (
	"context"
	"fmt"
	"log"

	"github.com/google/glazier/go/power"
)

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
