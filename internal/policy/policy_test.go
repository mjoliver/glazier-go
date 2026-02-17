package policy

import (
	"runtime"
	"testing"
)

func TestOSVersionPolicy_Check(t *testing.T) {
	tests := []struct {
		name    string
		policy  OSVersionPolicy
		wantErr bool
	}{
		{"no constraints - passes", OSVersionPolicy{}, false},
		{"correct OS", OSVersionPolicy{OS: runtime.GOOS}, false},
		{"wrong OS", OSVersionPolicy{OS: "fakeos"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Check()
			if (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOSVersionPolicy_ExactVersionMatch(t *testing.T) {
	current := getCurrentWindowsVersion()
	t.Logf("Current version: %q", current)

	// Should pass with current version in list
	p := &OSVersionPolicy{AllowedVersions: []string{current}}
	if err := p.Check(); err != nil {
		t.Errorf("should pass with matching version, got: %v", err)
	}

	// Should pass with current version among many
	p = &OSVersionPolicy{AllowedVersions: []string{"99", current, "98"}}
	if err := p.Check(); err != nil {
		t.Errorf("should pass when current is in list, got: %v", err)
	}

	// Should fail with non-matching version
	p = &OSVersionPolicy{AllowedVersions: []string{"NonExistentVersion"}}
	if err := p.Check(); err == nil {
		t.Error("should fail with non-matching version")
	}
}

func TestOSVersionPolicy_ServerVersions(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}

	current := getCurrentWindowsVersion()
	t.Logf("Detected Windows version: %q", current)

	// Verify that a Server 2019 policy does NOT match a Windows 11 host
	if current == "11" {
		p := &OSVersionPolicy{AllowedVersions: []string{"Server 2019"}}
		if err := p.Check(); err == nil {
			t.Error("Windows 11 should NOT match Server 2019 policy")
		}
	}

	// Verify that a Windows 10 policy does NOT match a Windows 11 host
	if current == "11" {
		p := &OSVersionPolicy{AllowedVersions: []string{"10"}}
		if err := p.Check(); err == nil {
			t.Error("Windows 11 should NOT match Windows 10 policy")
		}
	}
}

func TestWindowsVersion(t *testing.T) {
	ver := WindowsVersion()
	t.Logf("WindowsVersion(): %s", ver)
}

func TestDeviceModelPolicy_Check(t *testing.T) {
	p := &DeviceModelPolicy{}
	if err := p.Check(); err != nil {
		t.Errorf("empty list should pass, got: %v", err)
	}

	model := getDeviceModel()
	t.Logf("Detected model: %q", model)

	if model != "" {
		p = &DeviceModelPolicy{AllowedModels: []string{model}}
		if err := p.Check(); err != nil {
			t.Errorf("matching model should pass, got: %v", err)
		}

		p = &DeviceModelPolicy{AllowedModels: []string{"NonExistent12345"}}
		if err := p.Check(); err == nil {
			t.Error("non-matching model should fail")
		}
	}
}

func TestChassisTypePolicy_Check(t *testing.T) {
	p := &ChassisTypePolicy{}
	if err := p.Check(); err != nil {
		t.Errorf("empty list should pass, got: %v", err)
	}

	chassis := getChassisType()
	t.Logf("Detected chassis: %q", chassis)

	if chassis != "" {
		p = &ChassisTypePolicy{AllowedTypes: []string{chassis}}
		if err := p.Check(); err != nil {
			t.Errorf("matching chassis should pass, got: %v", err)
		}
	}
}

func TestNewPolicy_WithConfig(t *testing.T) {
	// Single version
	p, err := NewPolicy("os_version", map[string]interface{}{
		"os":      runtime.GOOS,
		"version": getCurrentWindowsVersion(),
	})
	if err != nil {
		t.Fatalf("NewPolicy() error = %v", err)
	}
	if err := p.Check(); err != nil {
		t.Errorf("should pass, got: %v", err)
	}

	// Multiple versions
	p, err = NewPolicy("os_version", map[string]interface{}{
		"allowed_versions": []interface{}{"10", "11", "Server 2022"},
	})
	if err != nil {
		t.Fatalf("NewPolicy() error = %v", err)
	}

	// Unknown policy
	_, err = NewPolicy("unknown", nil)
	if err == nil {
		t.Error("should fail for unknown policy")
	}
}
