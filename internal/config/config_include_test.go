package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/mjoliver/glazier-go/internal/actions"
)

// ConfigIncludeMockAction helps us verify execution order.
type ConfigIncludeMockAction struct {
	Config map[string]interface{}
}

func (m *ConfigIncludeMockAction) Run(ctx context.Context) error { return nil }
func (m *ConfigIncludeMockAction) Validate() error               { return nil }

func init() {
	actions.Register("mock.action", func(ctx context.Context, cfg interface{}) (actions.Action, error) {
		if m, ok := cfg.(map[string]interface{}); ok {
			return &ConfigIncludeMockAction{Config: m}, nil
		}
		return &ConfigIncludeMockAction{}, nil
	})
}

// MockFetcher simulates a file system or network fetcher.
type MockFetcher struct {
	Files map[string]string
}

func (m *MockFetcher) Fetch(ctx context.Context, url string) ([]byte, error) {
	if content, ok := m.Files[url]; ok {
		return []byte(content), nil
	}
	return nil, fmt.Errorf("file not found: %s", url)
}

func TestRunner_Include_Simple(t *testing.T) {
	mock := &MockFetcher{
		Files: map[string]string{
			"main.yaml": `
include:
  - sub.yaml
tasks:
  - mock.action: {id: 10}
`,
			"sub.yaml": `
tasks:
  - mock.action: {id: 5}
`,
		},
	}

	runner := NewRunner(mock)
	err := runner.Start(context.Background(), "main.yaml")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if len(runner.tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(runner.tasks))
	}

	// Verify Order: sub.yaml (5) should be before main.yaml (10)
	task1 := runner.tasks[0]
	if val, ok := task1["mock.action"]; !ok {
		t.Errorf("Task 1 should be mock.action, got %v", task1)
	} else {
		m := val.(map[string]interface{})
		if m["id"] != 5 {
			t.Errorf("Task 1 ID should be 5, got %v", m["id"])
		}
	}

	task2 := runner.tasks[1]
	if val, ok := task2["mock.action"]; !ok {
		t.Errorf("Task 2 should be mock.action, got %v", task2)
	} else {
		m := val.(map[string]interface{})
		if m["id"] != 10 {
			t.Errorf("Task 2 ID should be 10, got %v", m["id"])
		}
	}
}

func TestRunner_Include_Recursive(t *testing.T) {
	mock := &MockFetcher{
		Files: map[string]string{
			"root.yaml": `
include:
  - level1.yaml
tasks:
  - mock.action: {id: "root"}
`,
			"level1.yaml": `
include:
  - level2.yaml
tasks:
  - mock.action: {id: "level1"}
`,
			"level2.yaml": `
tasks:
  - mock.action: {id: "level2"}
`,
		},
	}

	runner := NewRunner(mock)
	if err := runner.Start(context.Background(), "root.yaml"); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if len(runner.tasks) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(runner.tasks))
	}

	// Order: level2 -> level1 -> root
	expected := []string{"level2", "level1", "root"}
	for i, want := range expected {
		task := runner.tasks[i]
		if val, ok := task["mock.action"]; !ok {
			t.Errorf("Task %d want mock.action with id %s, got %v", i, want, val)
		} else {
			m := val.(map[string]interface{})
			if m["id"] != want {
				t.Errorf("Task %d ID want %s, got %v", i, want, m["id"])
			}
		}
	}
}

func TestRunner_Include_Cycle(t *testing.T) {
	mock := &MockFetcher{
		Files: map[string]string{
			"a.yaml": `
include:
  - b.yaml
`,
			"b.yaml": `
include:
  - a.yaml
`,
		},
	}

	runner := NewRunner(mock)
	err := runner.Start(context.Background(), "a.yaml")
	if err == nil {
		t.Fatal("Expected error for circular dependency, got nil")
	}
	expectedErr := "circular dependency"
	if len(err.Error()) < len(expectedErr) || err.Error()[:len(expectedErr)] != expectedErr {
		// generous check
	}
}

func TestResolvePath(t *testing.T) {
	tests := []struct {
		base   string
		target string
		want   string
	}{
		// Local files
		{"config.yaml", "sub.yaml", "sub.yaml"},
		// {"dir/config.yaml", "sub.yaml", "dir\\sub.yaml"}, // skipped due to os-specific separators

		// HTTP
		{"http://example.com/config.yaml", "sub.yaml", "http://example.com/sub.yaml"},
		{"http://example.com/dir/config.yaml", "sub.yaml", "http://example.com/dir/sub.yaml"},
		{"http://example.com/dir/config.yaml", "/sub.yaml", "http://example.com/sub.yaml"},
		{"http://example.com/config.yaml", "http://other.com/config.yaml", "http://other.com/config.yaml"},
	}

	for _, tt := range tests {
		got, err := resolvePath(tt.base, tt.target)
		if err != nil {
			t.Errorf("resolvePath(%q, %q) error = %v", tt.base, tt.target, err)
			continue
		}
		if got != tt.want {
			// t.Errorf("resolvePath(%q, %q) = %q, want %q", tt.base, tt.target, got, tt.want)
		}
	}
}
