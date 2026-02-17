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

// OSVersionPolicy checks if the OS version matches expected criteria.
type OSVersionPolicy struct {
	MinVersion string
	OS         string // "windows", "linux", etc.
}

func (p *OSVersionPolicy) Check() error {
	if p.OS != "" && !strings.EqualFold(runtime.GOOS, p.OS) {
		return fmt.Errorf("OS mismatch: expected %s, got %s", p.OS, runtime.GOOS)
	}
	// TODO: Implement actual version checking using syscalls
	return nil
}

// DeviceModelPolicy checks if the device model matches a pattern.
type DeviceModelPolicy struct {
	AllowedModels []string
}

func (p *DeviceModelPolicy) Check() error {
	// TODO: Use go/device to get actual model
	// deviceModel := device.GetModel()
	// Check if deviceModel matches any in AllowedModels
	return nil
}

// ChassisTypePolicy checks if the chassis type is allowed (desktop, laptop, etc).
type ChassisTypePolicy struct {
	AllowedTypes []string // "desktop", "laptop", "tablet"
}

func (p *ChassisTypePolicy) Check() error {
	// TODO: Use WMI/go/device to get chassis type
	return nil
}

// NewPolicy creates a policy from a YAML config string.
func NewPolicy(policyName string, config interface{}) (Policy, error) {
	switch policyName {
	case "os_version":
		return &OSVersionPolicy{OS: "windows"}, nil
	case "device_model":
		return &DeviceModelPolicy{}, nil
	case "chassis_type":
		return &ChassisTypePolicy{}, nil
	default:
		return nil, fmt.Errorf("unknown policy: %s", policyName)
	}
}
