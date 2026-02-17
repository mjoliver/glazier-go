package actions

import (
	"context"
	"testing"
)

func TestBitLockerEnable_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  BitLockerEnableConfig
		wantErr bool
	}{
		{
			name:    "valid config with mode",
			config:  BitLockerEnableConfig{Mode: "tpm", Backup: true},
			wantErr: false,
		},
		{
			name:    "missing mode",
			config:  BitLockerEnableConfig{Backup: true},
			wantErr: true,
		},
		{
			name:    "valid config without backup",
			config:  BitLockerEnableConfig{Mode: "password", Backup: false},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &BitLockerEnable{Config: tt.config}
			t.Logf("Testing BitLocker config: mode=%q, backup=%v", tt.config.Mode, tt.config.Backup)

			err := action.Validate()

			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			} else if err == nil {
				t.Logf("✓ Validation passed")
			}
		})
	}
}

func TestGooGetInstall_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  GooGetInstallConfig
		wantErr bool
	}{
		{
			name:    "valid config with packages",
			config:  GooGetInstallConfig{Packages: []string{"chrome", "firefox"}},
			wantErr: false,
		},
		{
			name:    "empty package list",
			config:  GooGetInstallConfig{Packages: []string{}},
			wantErr: true,
		},
		{
			name:    "valid with reinstall flag",
			config:  GooGetInstallConfig{Packages: []string{"chrome"}, Reinstall: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &GooGetInstall{Config: tt.config}
			err := action.Validate()

			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestDomainJoin_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  DomainJoinConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: DomainJoinConfig{
				Domain:   "example.com",
				OU:       "OU=Computers,DC=example,DC=com",
				User:     "admin",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name:    "missing domain",
			config:  DomainJoinConfig{OU: "OU=Computers"},
			wantErr: true,
		},
		{
			name:    "domain only (minimal)",
			config:  DomainJoinConfig{Domain: "example.com"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &DomainJoin{Config: tt.config}
			err := action.Validate()

			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestPower_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  PowerConfig
		wantErr bool
	}{
		{
			name:    "valid reboot",
			config:  PowerConfig{Type: "reboot", Delay: 30},
			wantErr: false,
		},
		{
			name:    "valid shutdown",
			config:  PowerConfig{Type: "shutdown", Delay: 60},
			wantErr: false,
		},
		{
			name:    "invalid type",
			config:  PowerConfig{Type: "hibernate"},
			wantErr: true,
		},
		{
			name:    "empty type",
			config:  PowerConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &Power{Config: tt.config}
			err := action.Validate()

			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestNewGooGetInstall_ShorthandSyntax(t *testing.T) {
	// Test the shorthand YAML syntax: ["chrome", "firefox"]
	yamlData := []interface{}{"chrome", "firefox"}

	action, err := NewGooGetInstall(context.Background(), yamlData)
	if err != nil {
		t.Fatalf("NewGooGetInstall() unexpected error = %v", err)
	}

	ggAction, ok := action.(*GooGetInstall)
	if !ok {
		t.Fatal("NewGooGetInstall() did not return *GooGetInstall")
	}

	if len(ggAction.Config.Packages) != 2 {
		t.Errorf("NewGooGetInstall() got %d packages, want 2", len(ggAction.Config.Packages))
	}

	if ggAction.Config.Packages[0] != "chrome" || ggAction.Config.Packages[1] != "firefox" {
		t.Errorf("NewGooGetInstall() packages = %v, want [chrome, firefox]", ggAction.Config.Packages)
	}
}
