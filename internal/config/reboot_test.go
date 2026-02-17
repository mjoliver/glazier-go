package config

import (
	"context"
	"testing"

	"github.com/mjoliver/glazier-go/internal/actions"
)

// RebootMockAction helps us verify execution order and count.
type RebootMockAction struct {
	Count *int
}

func (m *RebootMockAction) Run(ctx context.Context) error {
	*m.Count++
	return nil
}

func (m *RebootMockAction) Validate() error { return nil }

func init() {
	actions.Register("reboot.check", func(ctx context.Context, cfg interface{}) (actions.Action, error) {
		// In a real scenario we'd parse config, but here we just need a handle to the counter
		// which we can't easily pass through standard factory without a global or similar hack for tests.
		// So we'll use a slightly different approach: the factory returns a struct,
		// but since we need to assertions on the SAME instance or shared state,
		// we typically use a channel or closure.
		//
		// Simplified for this test: We'll use a global map for test state if needed,
		// or just define the factory INSIDE the test if registry allows overwriting (it panics on dupe).
		//
		// Since Registry panics on duplicate registration, and we can't easily unregister,
		// we'll use a unique name for this test or rely on the fact that `init` runs once.
		// But we need to reset state between tests.
		//
		// Better approach: usage of a shared counter in the factory closure.
		return nil, nil // placeholder, actual implementation in test
	})
}

// Global state for invalidation/counting in tests (mutex omitted for simple test)
var executionCount int

func TestRunner_Reboot_Repetition(t *testing.T) {
	// 1. Setup a unique action for this test to avoid collision
	actionName := "reboot.check.repetition"

	// We need to register it safely. If it's already there (from previous run), we can't re-register.
	// But `actions.Registry` is a map, we can manually check or recover.
	// However, the `actions` package doesn't expose a way to unregister.
	// We will rely on a new unique name or just use a standard mock if available.
	// Let's rely on a global variable for counting that we reset.

	executionCount = 0

	// Optimistic registration (might panic if run twice in same process, but go test usually isolates)
	// To be safe in `go test` filtering, we check if it is already registered.
	if _, exists := actions.Registry[actionName]; !exists {
		actions.Register(actionName, func(ctx context.Context, cfg interface{}) (actions.Action, error) {
			return &RebootMockAction{Count: &executionCount}, nil
		})
	}

	mockConfig := `
tasks:
  - reboot.check.repetition: {}
`

	mockFetcher := &MockFetcher{
		Files: map[string]string{
			"reboot.yaml": mockConfig,
		},
	}

	// 2. First Run (Boot 1)
	runner1 := NewRunner(mockFetcher)
	if err := runner1.Start(context.Background(), "reboot.yaml"); err != nil {
		t.Fatalf("First Run failed: %v", err)
	}

	if executionCount != 1 {
		t.Errorf("Expected execution count 1 after first run, got %d", executionCount)
	}

	// 3. Second Run (Boot 2 / Reboot)
	// We create a new runner to simulate a restart, but using the same fetcher/config source.
	runner2 := NewRunner(mockFetcher)
	if err := runner2.Start(context.Background(), "reboot.yaml"); err != nil {
		t.Fatalf("Second Run failed: %v", err)
	}

	// 4. Verify Repetition
	// Since there is no persistence, the task should run AGAIN.
	// Total count should be 2.
	if executionCount != 2 {
		t.Errorf("Expected execution count 2 after second run (proving repetition), got %d. Task persistence might strictly be preventing re-run!", executionCount)
	}
}
