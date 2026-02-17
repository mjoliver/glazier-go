package actions

import (
	"context"
	"testing"
)

func TestRegistrySet_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  RegistryConfig
		wantErr bool
	}{
		{"valid", RegistryConfig{Path: `SOFTWARE\Glazier`, Name: "Version", Value: "1.0"}, false},
		{"missing path", RegistryConfig{Name: "Version"}, true},
		{"missing name", RegistryConfig{Path: `SOFTWARE\Glazier`}, true},
		{"both missing", RegistryConfig{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &RegistrySet{Config: tt.config}
			err := a.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistryDelete_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  RegistryConfig
		wantErr bool
	}{
		{"valid", RegistryConfig{Path: `SOFTWARE\Glazier`, Name: "Version"}, false},
		{"missing path", RegistryConfig{Name: "Version"}, true},
		{"missing name", RegistryConfig{Path: `SOFTWARE\Glazier`}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &RegistryDelete{Config: tt.config}
			err := a.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistryGet_Validate(t *testing.T) {
	a := &RegistryGet{Config: RegistryConfig{Path: `SOFTWARE\Glazier`, Name: "Version"}}
	if err := a.Validate(); err != nil {
		t.Errorf("Validate() should pass, got: %v", err)
	}

	a = &RegistryGet{Config: RegistryConfig{}}
	if err := a.Validate(); err == nil {
		t.Error("Validate() should fail with empty config")
	}
}

func TestRegistrySet_DefaultType(t *testing.T) {
	a := &RegistrySet{Config: RegistryConfig{Path: `SOFTWARE\Test`, Name: "Key"}}
	a.Validate()
	if a.Config.Type != "string" {
		t.Errorf("default type should be string, got %q", a.Config.Type)
	}
}

func TestRegistryActions_Registration(t *testing.T) {
	actions := []string{"registry.set", "registry.delete", "registry.get"}
	for _, name := range actions {
		t.Run(name, func(t *testing.T) {
			_, err := New(context.Background(), name, map[string]interface{}{
				"path": `SOFTWARE\Glazier`,
				"name": "TestValue",
			})
			if err != nil {
				t.Errorf("New(%q) error = %v", name, err)
			}
		})
	}
}
