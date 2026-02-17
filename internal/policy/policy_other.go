//go:build !windows

package policy

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// getOSVersion returns the OS major, minor, and build version on non-Windows.
func getOSVersion() (int, int, int) {
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("sw_vers", "-productVersion").Output()
		if err == nil {
			parts := strings.SplitN(strings.TrimSpace(string(out)), ".", 3)
			major, _ := strconv.Atoi(parts[0])
			minor := 0
			if len(parts) > 1 {
				minor, _ = strconv.Atoi(parts[1])
			}
			build := 0
			if len(parts) > 2 {
				build, _ = strconv.Atoi(parts[2])
			}
			return major, minor, build
		}
	} else if runtime.GOOS == "linux" {
		out, err := exec.Command("uname", "-r").Output()
		if err == nil {
			parts := strings.SplitN(strings.TrimSpace(string(out)), ".", 3)
			major, _ := strconv.Atoi(parts[0])
			minor := 0
			if len(parts) > 1 {
				minor, _ = strconv.Atoi(parts[1])
			}
			return major, minor, 0
		}
	}
	return 0, 0, 0
}

// detectWindowsVersion is a no-op on non-Windows.
func detectWindowsVersion() string {
	return runtime.GOOS
}

func getDeviceModel() string {
	return ""
}

func getChassisType() string {
	return ""
}
