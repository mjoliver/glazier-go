package config

import (
	"context"
	"testing"

	"github.com/mjoliver/glazier-go/internal/actions"
)

func TestRunner_checkPolicy(t *testing.T) {
	tests := []struct {
		name        string
		policyData  interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid policy list",
			policyData: []interface{}{"os_version"},
			wantErr:    false,
		},
		{
			name:        "invalid policy format - not a list",
			policyData:  "os_version",
			wantErr:     true,
			errContains: "policy must be a list",
		},
		{
			name:       "empty policy list",
			policyData: []interface{}{},
			wantErr:    false,
		},
		{
			name:       "multiple policies",
			policyData: []interface{}{"os_version", "device_model"},
			wantErr:    false,
		},
		{
			name:        "unknown policy",
			policyData:  []interface{}{"unknown_policy"},
			wantErr:     true,
			errContains: "unknown policy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Runner{}
			err := r.checkPolicy(context.Background(), tt.policyData)

			if tt.wantErr {
				if err == nil {
					t.Errorf("checkPolicy() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("checkPolicy() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("checkPolicy() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestRunner_Start(t *testing.T) {
	// Register a test action
	testActionCalled := false
	actions.Register("test.action", func(ctx context.Context, data interface{}) (actions.Action, error) {
		return &testAction{called: &testActionCalled}, nil
	})

	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty task list",
			yamlContent: `[]`,
			wantErr:     false,
		},
		{
			name: "single action",
			yamlContent: `
- test.action: {}
`,
			wantErr: false,
		},
		{
			name: "unknown action",
			yamlContent: `
- unknown.action: {}
`,
			wantErr:     true,
			errContains: "unknown action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock fetcher
			fetcher := &mockFetcher{data: []byte(tt.yamlContent)}
			r := NewRunner(fetcher)

			err := r.Start(context.Background(), "test.yaml")

			if tt.wantErr {
				if err == nil {
					t.Errorf("Start() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Start() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Start() unexpected error = %v", err)
				}
			}
		})
	}
}

// Test helper types
type testAction struct {
	called *bool
}

func (a *testAction) Run(ctx context.Context) error {
	*a.called = true
	return nil
}

func (a *testAction) Validate() error {
	return nil
}

type mockFetcher struct {
	data []byte
	err  error
}

func (m *mockFetcher) Fetch(ctx context.Context, path string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data, nil
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
