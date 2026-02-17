//go:build windows

package policy

import (
	"fmt"
	"strings"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows"
)

// getOSVersion returns Windows major, minor, and build number.
func getOSVersion() (int, int, int) {
	ver := windows.RtlGetVersion()
	return int(ver.MajorVersion), int(ver.MinorVersion), int(ver.BuildNumber)
}

// WMI types
type win32OperatingSystem struct {
	Caption string
}

type win32ComputerSystem struct {
	Model string
}

type win32SystemEnclosure struct {
	ChassisTypes []int32
}

// detectWindowsVersion returns a user-facing version string: "11", "10",
// "Server 2022", "Server 2019", etc.
func detectWindowsVersion() string {
	// First try WMI for the most accurate caption (handles Server editions)
	var results []win32OperatingSystem
	if err := wmi.Query("SELECT Caption FROM Win32_OperatingSystem", &results); err == nil && len(results) > 0 {
		caption := results[0].Caption
		// Extract the version from captions like:
		//   "Microsoft Windows 11 Pro"
		//   "Microsoft Windows Server 2022 Standard"
		//   "Microsoft Windows 10 Enterprise"
		lower := strings.ToLower(caption)
		switch {
		case strings.Contains(lower, "server 2025"):
			return "Server 2025"
		case strings.Contains(lower, "server 2022"):
			return "Server 2022"
		case strings.Contains(lower, "server 2019"):
			return "Server 2019"
		case strings.Contains(lower, "server 2016"):
			return "Server 2016"
		case strings.Contains(lower, "windows 11"):
			return "11"
		case strings.Contains(lower, "windows 10"):
			return "10"
		}
	}

	// Fallback: use build number
	_, _, build := getOSVersion()
	if build >= 22000 {
		return "11"
	}
	return "10"
}

// getDeviceModel returns the device model string via WMI.
func getDeviceModel() string {
	var results []win32ComputerSystem
	if err := wmi.Query("SELECT Model FROM Win32_ComputerSystem", &results); err != nil {
		return ""
	}
	if len(results) > 0 {
		return strings.TrimSpace(results[0].Model)
	}
	return ""
}

// getChassisType returns a human-readable chassis type via WMI.
func getChassisType() string {
	var results []win32SystemEnclosure
	if err := wmi.Query("SELECT ChassisTypes FROM Win32_SystemEnclosure", &results); err != nil {
		return ""
	}
	if len(results) == 0 || len(results[0].ChassisTypes) == 0 {
		return ""
	}

	switch results[0].ChassisTypes[0] {
	case 1:
		return "other"
	case 2:
		return "unknown"
	case 3, 4, 5, 6, 7:
		return "desktop"
	case 8, 9, 10, 11, 12, 14, 18, 21:
		return "laptop"
	case 13:
		return "allinone"
	case 15, 16:
		return "server"
	case 17:
		return "docking"
	case 30, 31, 32:
		return "tablet"
	default:
		return fmt.Sprintf("type_%d", results[0].ChassisTypes[0])
	}
}
