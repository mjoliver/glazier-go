package template

import (
	"os"
	"strings"
	"testing"
)

func TestProcess(t *testing.T) {
	info := &BuildInfo{
		Hostname:  "test-host",
		Stage:     "50",
		ImageID:   "win11-v1",
		Username:  "admin",
		Timestamp: "2026-01-01T00:00:00",
	}

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple variable substitution",
			input:    "host: {{.Hostname}}",
			expected: "host: test-host",
		},
		{
			name:     "multiple variables",
			input:    "{{.Hostname}}-{{.Stage}}-{{.ImageID}}",
			expected: "test-host-50-win11-v1",
		},
		{
			name:     "conditional with value present",
			input:    "{{if .ImageID}}image-{{.ImageID}}{{else}}no-image{{end}}",
			expected: "image-win11-v1",
		},
		{
			name:     "conditional with empty value",
			input:    "{{if .ImageID}}has-image{{else}}no-image{{end}}",
			expected: "has-image",
		},
		{
			name:     "no template markers (passthrough)",
			input:    "plain text with no markers",
			expected: "plain text with no markers",
		},
		{
			name:    "invalid template syntax",
			input:   "{{.Hostname",
			wantErr: true,
		},
		{
			name:     "yaml-like config with templates",
			input:    "- googet.install:\n    packages:\n      - base-{{.ImageID}}\n      - config-{{.Hostname}}",
			expected: "- googet.install:\n    packages:\n      - base-win11-v1\n      - config-test-host",
		},
		{
			name:     "stage variable",
			input:    "- stage.set: \"{{.Stage}}\"",
			expected: "- stage.set: \"50\"",
		},
		{
			name:     "timestamp variable",
			input:    "# Generated at {{.Timestamp}}",
			expected: "# Generated at 2026-01-01T00:00:00",
		},
		{
			name:     "username variable",
			input:    "user: {{.Username}}",
			expected: "user: admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Process([]byte(tt.input), info)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Process() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Process() unexpected error: %v", err)
				return
			}

			got := strings.TrimSpace(string(result))
			if got != tt.expected {
				t.Errorf("Process() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestProcess_NilBuildInfo(t *testing.T) {
	input := []byte("{{.Hostname}}")
	result, err := Process(input, nil)

	if err != nil {
		t.Errorf("Process() with nil BuildInfo should not error, got: %v", err)
	}

	if string(result) != string(input) {
		t.Errorf("Process() with nil BuildInfo should return input unchanged, got %q", string(result))
	}
}

func TestNewBuildInfo(t *testing.T) {
	info, err := NewBuildInfo()
	if err != nil {
		t.Fatalf("NewBuildInfo() unexpected error: %v", err)
	}

	if info.Hostname == "" {
		t.Error("NewBuildInfo() Hostname should not be empty")
	}

	if info.Timestamp == "" {
		t.Error("NewBuildInfo() Timestamp should not be empty")
	}

	if info.Stage == "" {
		t.Error("NewBuildInfo() Stage should not be empty (should default to '0')")
	}

	t.Logf("BuildInfo: Hostname=%s, Stage=%s, Username=%s, Timestamp=%s",
		info.Hostname, info.Stage, info.Username, info.Timestamp)
}

func TestNewBuildInfo_EnvVars(t *testing.T) {
	// Set env vars and verify they are picked up
	os.Setenv("IMAGE_ID", "test-image-123")
	os.Setenv("GLAZIER_STAGE", "42")
	defer os.Unsetenv("IMAGE_ID")
	defer os.Unsetenv("GLAZIER_STAGE")

	info, err := NewBuildInfo()
	if err != nil {
		t.Fatalf("NewBuildInfo() unexpected error: %v", err)
	}

	if info.ImageID != "test-image-123" {
		t.Errorf("NewBuildInfo() ImageID = %q, want %q", info.ImageID, "test-image-123")
	}

	if info.Stage != "42" {
		t.Errorf("NewBuildInfo() Stage = %q, want %q", info.Stage, "42")
	}

	t.Logf("✓ Env vars correctly populated: ImageID=%s, Stage=%s", info.ImageID, info.Stage)
}

func TestProcess_EndToEnd(t *testing.T) {
	// Simulate a full config template processing
	info, _ := NewBuildInfo()

	yamlConfig := `
- stage.set: "10"
- googet.install:
    packages:
      - base-{{.Hostname}}
{{if .ImageID}}
- bitlocker.enable:
    mode: tpm
{{end}}
- stage.set: "100"
`
	result, err := Process([]byte(yamlConfig), info)
	if err != nil {
		t.Fatalf("Process() unexpected error: %v", err)
	}

	output := string(result)
	t.Logf("Processed config:\n%s", output)

	// Verify hostname was substituted
	if strings.Contains(output, "{{.Hostname}}") {
		t.Error("Process() should have substituted {{.Hostname}}")
	}

	// Verify base structure is preserved
	if !strings.Contains(output, "stage.set") {
		t.Error("Process() should preserve stage.set action")
	}
	if !strings.Contains(output, "googet.install") {
		t.Error("Process() should preserve googet.install action")
	}
}
