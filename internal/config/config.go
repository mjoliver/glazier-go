package config

import (
	"context"
	"fmt"
	"log"

	"gopkg.in/yaml.v3"

	"github.com/google/glazier/internal/actions"
	"github.com/google/glazier/internal/policy"
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
		log.Printf("Executing task %d/%d", i+1, len(r.tasks))

		for key, val := range task {
			if key == "policy" {
				// Handle policy check
				if err := r.checkPolicy(ctx, val); err != nil {
					return fmt.Errorf("policy check failed: %w", err)
				}
				continue
			}

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

			if err := action.Run(ctx); err != nil {
				return fmt.Errorf("action %s execution failed: %w", key, err)
			}
		}
	}
	return nil
}

func (r *Runner) checkPolicy(ctx context.Context, policyData interface{}) error {
	// policyData can be a list of policy names
	policies, ok := policyData.([]interface{})
	if !ok {
		return fmt.Errorf("policy must be a list")
	}

	for _, p := range policies {
		policyName, ok := p.(string)
		if !ok {
			log.Printf("Warning: invalid policy format: %v", p)
			continue
		}

		log.Printf("Checking policy: %s", policyName)

		// Create and check the policy
		policy, err := policy.NewPolicy(policyName, nil)
		if err != nil {
			return fmt.Errorf("failed to create policy %s: %w", policyName, err)
		}

		if err := policy.Check(); err != nil {
			return fmt.Errorf("policy check failed for %s: %w", policyName, err)
		}

		log.Printf("Policy %s passed", policyName)
	}

	return nil
}
