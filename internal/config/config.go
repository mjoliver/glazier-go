package config

import (
	"context"
	"fmt"
	"time"

	"github.com/google/deck"
	"gopkg.in/yaml.v3"

	"github.com/mjoliver/glazier-go/internal/actions"
	"github.com/mjoliver/glazier-go/internal/policy"
)

// TaskList represents the top-level YAML structure of a Glazier config.
// It is a list of map items, where each item is either a "policy" check or an action.
type TaskList []map[string]interface{}

// Runner executes a task list.
type Runner struct {
	tasks   TaskList
	fetcher FetcherInterface
}

// NewRunner creates a new Runner.
func NewRunner(f FetcherInterface) *Runner {
	return &Runner{
		fetcher: f,
	}
}

// Start executes the task list processing, starting from the given config path.
func (r *Runner) Start(ctx context.Context, configURL string) error {
	data, err := r.fetcher.Fetch(ctx, configURL)
	if err != nil {
		return fmt.Errorf("failed to fetch initial config: %w", err)
	}

	if err := yaml.Unmarshal(data, &r.tasks); err != nil {
		return fmt.Errorf("failed to parse task list: %w", err)
	}

	for i, task := range r.tasks {
		deck.Infof("Executing task %d/%d", i+1, len(r.tasks))

		for key, val := range task {
			if key == "policy" {
				// Handle policy check
				if err := r.checkPolicy(ctx, val); err != nil {
					return fmt.Errorf("policy check failed: %w", err)
				}
				continue
			}

			// Extract retry/error config before passing to factory
			retries, onError := extractRunOpts(val)

			// Handle action
			factory, ok := actions.Registry[key]
			if !ok {
				return fmt.Errorf("unknown action: %s", key)
			}

			action, err := factory(ctx, val)
			if err != nil {
				return fmt.Errorf("failed to create action %s: %w", key, err)
			}

			if err := action.Validate(); err != nil {
				return fmt.Errorf("action %s validation failed: %w", key, err)
			}

			if err := runWithRetry(ctx, key, action, retries); err != nil {
				if onError == "continue" {
					deck.Warningf("Action %s failed (continuing): %v", key, err)
					continue
				}
				return fmt.Errorf("action %s execution failed: %w", key, err)
			}
		}
	}
	return nil
}

// extractRunOpts pulls retries and on_error from action config if present.
func extractRunOpts(val interface{}) (int, string) {
	m, ok := val.(map[string]interface{})
	if !ok {
		return 0, ""
	}
	retries := 0
	if r, ok := m["retries"]; ok {
		switch v := r.(type) {
		case int:
			retries = v
		case float64:
			retries = int(v)
		}
	}
	onError := ""
	if e, ok := m["on_error"]; ok {
		if s, ok := e.(string); ok {
			onError = s
		}
	}
	return retries, onError
}

// runWithRetry executes an action with exponential backoff retries.
func runWithRetry(ctx context.Context, name string, action actions.Action, retries int) error {
	var lastErr error
	attempts := retries + 1 // retries=0 means 1 attempt

	for attempt := 1; attempt <= attempts; attempt++ {
		lastErr = action.Run(ctx)
		if lastErr == nil {
			return nil
		}

		if attempt < attempts {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second // 1s, 2s, 4s, 8s...
			deck.Warningf("Action %s attempt %d/%d failed: %v (retrying in %v)", name, attempt, attempts, lastErr, backoff)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	return lastErr
}

func (r *Runner) checkPolicy(ctx context.Context, policyData interface{}) error {
	// policyData can be a list of policy names or policy maps
	policies, ok := policyData.([]interface{})
	if !ok {
		return fmt.Errorf("policy must be a list")
	}

	for _, p := range policies {
		var policyName string
		var policyConfig interface{}

		switch v := p.(type) {
		case string:
			// Bare policy name: - os_version
			policyName = v
		case map[string]interface{}:
			// Policy with config: - os_version: {version: "11"}
			for k, val := range v {
				policyName = k
				policyConfig = val
				break // Only one key per map entry
			}
		default:
			deck.Warningf("Invalid policy format: %v", p)
			continue
		}

		deck.Infof("Checking policy: %s", policyName)

		pol, err := policy.NewPolicy(policyName, policyConfig)
		if err != nil {
			return fmt.Errorf("failed to create policy %s: %w", policyName, err)
		}

		if err := pol.Check(); err != nil {
			return fmt.Errorf("policy check failed for %s: %w", policyName, err)
		}

		deck.Infof("Policy %s passed", policyName)
	}

	return nil
}
