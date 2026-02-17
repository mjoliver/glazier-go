package policy

import (
	"fmt"
	"runtime"
	"strings"
)

// Policy represents a system requirement check.
type Policy interface {
	// Check evaluates the policy and returns an error if it fails.
	Check() error
}

// OSVersionPolicy checks if the OS and version match expected criteria.
// Uses exact version matching to prevent configs meant for one OS version
// from accidentally running on a different version (e.g. Server 2019 vs 2022).
type OSVersionPolicy struct {
	OS              string   // "windows", "linux", etc.
	AllowedVersions []string // e.g. ["10", "11"] or ["2019", "2022"]
}

func (p *OSVersionPolicy) Check() error {
	// Check OS type
	if p.OS != "" && !strings.EqualFold(runtime.GOOS, p.OS) {
		return fmt.Errorf("policy os_version: expected OS %q, got %q", p.OS, runtime.GOOS)
	}

	// Check version if specified
	if len(p.AllowedVersions) > 0 {
		current := getCurrentWindowsVersion()

		for _, allowed := range p.AllowedVersions {
			if strings.EqualFold(current, allowed) {
				return nil
			}
		}

		return fmt.Errorf("policy os_version: version %q not in allowed list %v", current, p.AllowedVersions)
	}

	return nil
}

// currentWindowsVersion returns the user-facing Windows version string.
// On Windows: "11", "10", "Server 2022", "Server 2019", etc.
// On other OS: runtime.GOOS
// getCurrentWindowsVersion returns the user-facing Windows version string.
// On Windows: "11", "10", "Server 2022", "Server 2019", etc.
// On other OS: runtime.GOOS
var getCurrentWindowsVersion = func() string {
	if runtime.GOOS != "windows" {
		return runtime.GOOS
	}
	return detectWindowsVersion()
}

// WindowsVersion returns a human-friendly Windows version for logging.
func WindowsVersion() string {
	if runtime.GOOS != "windows" {
		return runtime.GOOS
	}
	_, _, build := getOSVersion()
	ver := detectWindowsVersion()
	return fmt.Sprintf("Windows %s (build %d)", ver, build)
}

// DeviceModelPolicy checks if the device model matches allowed models.
type DeviceModelPolicy struct {
	AllowedModels []string
}

func (p *DeviceModelPolicy) Check() error {
	if len(p.AllowedModels) == 0 {
		return nil
	}

	model := getDeviceModel()
	if model == "" {
		return nil
	}

	for _, allowed := range p.AllowedModels {
		if strings.Contains(strings.ToLower(model), strings.ToLower(allowed)) {
			return nil
		}
	}

	return fmt.Errorf("policy device_model: model %q not in allowed list %v", model, p.AllowedModels)
}

// ChassisTypePolicy checks if the chassis type is allowed.
type ChassisTypePolicy struct {
	AllowedTypes []string // "desktop", "laptop", "tablet", "server"
}

func (p *ChassisTypePolicy) Check() error {
	if len(p.AllowedTypes) == 0 {
		return nil
	}

	chassis := getChassisType()
	if chassis == "" {
		return nil
	}

	for _, allowed := range p.AllowedTypes {
		if strings.EqualFold(chassis, allowed) {
			return nil
		}
	}

	return fmt.Errorf("policy chassis_type: type %q not in allowed list %v", chassis, p.AllowedTypes)
}

// NewPolicy creates a policy from a YAML config.
func NewPolicy(policyName string, config interface{}) (Policy, error) {
	switch policyName {
	case "os_version":
		p := &OSVersionPolicy{OS: "windows"}
		if m, ok := config.(map[string]interface{}); ok {
			if v, ok := m["os"].(string); ok {
				p.OS = v
			}
			if v, ok := m["version"].(string); ok {
				p.AllowedVersions = []string{v}
			}
			if v, ok := m["allowed_versions"].([]interface{}); ok {
				for _, item := range v {
					if s, ok := item.(string); ok {
						p.AllowedVersions = append(p.AllowedVersions, s)
					}
				}
			}
		}
		return p, nil
	case "device_model":
		p := &DeviceModelPolicy{}
		if m, ok := config.(map[string]interface{}); ok {
			if v, ok := m["allowed"].([]interface{}); ok {
				for _, item := range v {
					if s, ok := item.(string); ok {
						p.AllowedModels = append(p.AllowedModels, s)
					}
				}
			}
		}
		return p, nil
	case "chassis_type":
		p := &ChassisTypePolicy{}
		if m, ok := config.(map[string]interface{}); ok {
			if v, ok := m["allowed"].([]interface{}); ok {
				for _, item := range v {
					if s, ok := item.(string); ok {
						p.AllowedTypes = append(p.AllowedTypes, s)
					}
				}
			}
		}
		return p, nil
	default:
		return nil, fmt.Errorf("unknown policy: %s", policyName)
	}
}
