package config

import (
	"context"
	"fmt"

	"github.com/google/deck"
	"github.com/mjoliver/glazier-go/internal/actions"
	"github.com/mjoliver/glazier-go/internal/policy"
)

// Validate checks a TaskList for errors without executing actions.
func Validate(ctx context.Context, tasks TaskList) error {
	errorCount := 0

	for i, task := range tasks {
		for key, val := range task {
			if key == "policy" {
				if err := validatePolicy(val); err != nil {
					deck.Errorf("Task %d [Policy] Invalid: %v", i+1, err)
					errorCount++
				} else {
					deck.Infof("Task %d [Policy] OK", i+1)
				}
				continue
			}

			// Action Check
			factory, ok := actions.Registry[key]
			if !ok {
				deck.Errorf("Task %d [Action %s] Unknown action type", i+1, key)
				errorCount++
				continue
			}

			action, err := factory(ctx, val)
			if err != nil {
				deck.Errorf("Task %d [Action %s] Validation Error (Factory): %v", i+1, key, err)
				errorCount++
				continue
			}

			if err := action.Validate(); err != nil {
				deck.Errorf("Task %d [Action %s] Validation Error: %v", i+1, key, err)
				errorCount++
			} else {
				deck.Infof("Task %d [Action %s] OK", i+1, key)
			}
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("validation failed with %d errors", errorCount)
	}
	return nil
}

func validatePolicy(policyData interface{}) error {
	policies, ok := policyData.([]interface{})
	if !ok {
		return fmt.Errorf("policy must be a list")
	}

	for _, p := range policies {
		var policyName string
		var policyConfig interface{}

		switch v := p.(type) {
		case string:
			policyName = v
		case map[string]interface{}:
			for k, val := range v {
				policyName = k
				policyConfig = val
				break
			}
		default:
			return fmt.Errorf("invalid policy format: %v", p)
		}

		// Try to create the policy to validate inputs
		_, err := policy.NewPolicy(policyName, policyConfig)
		if err != nil {
			return fmt.Errorf("invalid policy %s: %w", policyName, err)
		}
	}
	return nil
}
