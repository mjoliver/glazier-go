package policy

import (
	"testing"
)

func TestOSVersionPolicy_Check(t *testing.T) {
	tests := []struct {
		name    string
		policy  OSVersionPolicy
		wantErr bool
	}{
		{
			name:    "empty OS - should pass (no check)",
			policy:  OSVersionPolicy{OS: "", MinVersion: "10.0"},
			wantErr: false,
		},
		{
			name:    "windows OS on windows",
			policy:  OSVersionPolicy{OS: "windows", MinVersion: "10.0"},
			wantErr: false, // Will pass if running on Windows
		},
		{
			name:    "linux OS on windows",
			policy:  OSVersionPolicy{OS: "linux", MinVersion: "5.0"},
			wantErr: true, // Will fail if running on Windows
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Check()

			if tt.wantErr && err == nil {
				t.Errorf("OSVersionPolicy.Check() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("OSVersionPolicy.Check() unexpected error = %v", err)
			}
		})
	}
}

func TestDeviceModelPolicy_Check(t *testing.T) {
	tests := []struct {
		name    string
		policy  DeviceModelPolicy
		wantErr bool
	}{
		{
			name:    "empty allowed models",
			policy:  DeviceModelPolicy{AllowedModels: []string{}},
			wantErr: false, // Currently a no-op
		},
		{
			name:    "with allowed models",
			policy:  DeviceModelPolicy{AllowedModels: []string{"ThinkPad", "Latitude"}},
			wantErr: false, // Currently a no-op (needs go/device)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Check()

			if tt.wantErr && err == nil {
				t.Errorf("DeviceModelPolicy.Check() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("DeviceModelPolicy.Check() unexpected error = %v", err)
			}
		})
	}
}

func TestNewPolicy(t *testing.T) {
	tests := []struct {
		name       string
		policyName string
		config     interface{}
		wantErr    bool
		wantType   string
	}{
		{
			name:       "os_version policy",
			policyName: "os_version",
			config:     nil,
			wantErr:    false,
			wantType:   "*policy.OSVersionPolicy",
		},
		{
			name:       "device_model policy",
			policyName: "device_model",
			config:     nil,
			wantErr:    false,
			wantType:   "*policy.DeviceModelPolicy",
		},
		{
			name:       "chassis_type policy",
			policyName: "chassis_type",
			config:     nil,
			wantErr:    false,
			wantType:   "*policy.ChassisTypePolicy",
		},
		{
			name:       "unknown policy",
			policyName: "unknown_policy",
			config:     nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPolicy(tt.policyName, tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPolicy() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewPolicy() unexpected error = %v", err)
				return
			}

			if p == nil {
				t.Error("NewPolicy() returned nil policy")
			}
		})
	}
}
