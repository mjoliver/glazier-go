package config

import (
	"context"
	"fmt"
	"testing"
)

// MockAction for testing retries
type MockAction struct {
	FailCount int
	Calls     int
}

func (m *MockAction) Run(ctx context.Context) error {
	m.Calls++
	if m.Calls <= m.FailCount {
		return fmt.Errorf("mock failure %d", m.Calls)
	}
	return nil
}

func (m *MockAction) Validate() error { return nil }

func TestExtractRunOpts(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		wantRetries int
		wantError   string
	}{
		{
			name:        "nil input",
			input:       nil,
			wantRetries: 0,
			wantError:   "",
		},
		{
			name: "valid config",
			input: map[string]interface{}{
				"retries":  3,
				"on_error": "continue",
			},
			wantRetries: 3,
			wantError:   "continue",
		},
		{
			name: "float retries (yaml unmarshal)",
			input: map[string]interface{}{
				"retries": 5.0,
			},
			wantRetries: 5,
			wantError:   "",
		},
		{
			name: "partial config",
			input: map[string]interface{}{
				"retries": 2,
			},
			wantRetries: 2,
			wantError:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetries, gotError := extractRunOpts(tt.input)
			if gotRetries != tt.wantRetries {
				t.Errorf("extractRunOpts() retries = %v, want %v", gotRetries, tt.wantRetries)
			}
			if gotError != tt.wantError {
				t.Errorf("extractRunOpts() onError = %v, want %v", gotError, tt.wantError)
			}
		})
	}
}

func TestRunWithRetry(t *testing.T) {
	ctx := context.Background()

	t.Run("success first try", func(t *testing.T) {
		a := &MockAction{FailCount: 0}
		err := runWithRetry(ctx, "test", a, 3)
		if err != nil {
			t.Errorf("runWithRetry() unexpected error: %v", err)
		}
		if a.Calls != 1 {
			t.Errorf("runWithRetry() calls = %d, want 1", a.Calls)
		}
	})

	t.Run("retry and succeed", func(t *testing.T) {
		a := &MockAction{FailCount: 2} // Fails twice, succeeds 3rd time
		err := runWithRetry(ctx, "test", a, 3)
		if err != nil {
			t.Errorf("runWithRetry() unexpected error: %v", err)
		}
		if a.Calls != 3 {
			t.Errorf("runWithRetry() calls = %d, want 3", a.Calls)
		}
	})

	t.Run("retries exhausted", func(t *testing.T) {
		a := &MockAction{FailCount: 5}
		err := runWithRetry(ctx, "test", a, 2) // Retries=2 (total 3 attempts)
		if err == nil {
			t.Error("runWithRetry() expected error, got nil")
		}
		if a.Calls != 3 {
			t.Errorf("runWithRetry() calls = %d, want 3", a.Calls)
		}
	})
}
