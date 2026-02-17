package config

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/deck"
	"gopkg.in/yaml.v3"

	"github.com/mjoliver/glazier-go/internal/actions"
	"github.com/mjoliver/glazier-go/internal/policy"
)

// Config represents the schema of a configuration file.
type Config struct {
	Includes []string                 `yaml:"include"`
	Tasks    []map[string]interface{} `yaml:"tasks"`
	// Backwards compatibility: if the root is just a list, we handle that during unmarshal?
	// Actually, YAML v3 might struggle if we unmarshal a list into a struct.
	// We might need a custom UnmarshalYAML or try to unmarshal into []map first, then struct.
	// But to keep it clean, let's assume the NEW format is strict: root object with `tasks` list.
	// For backwards compat with the existing list-only format, we can try to detect or just support both.
	// Let's stick to the Plan: we add `include` which implies a change to struct-based config if used.
	// Wait, the existing config is a LIST of maps.
	// If we change to a STRUCT, old configs break.
	// Solution: Attempt to unmarshal as Config struct. If that implies empty tasks and includes, try unmarshalling as TaskList.
}

// ConfigFile is a helper to parse potential formats.
func parseConfigData(data []byte) (*Config, error) {
	// Try Unmarshalling as the new Struct format
	var c Config
	err := yaml.Unmarshal(data, &c)
	if err == nil && (len(c.Tasks) > 0 || len(c.Includes) > 0) {
		return &c, nil
	}

	// Fallback: Try Unmarshalling as the old List format
	var t TaskList
	if err := yaml.Unmarshal(data, &t); err == nil {
		// If it's a list, treat it as just tasks
		return &Config{Tasks: t}, nil
	}

	return nil, fmt.Errorf("failed to parse config as either struct or list")
}

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
	tasks, err := r.LoadConfig(ctx, configURL)
	if err != nil {
		return err
	}
	r.tasks = tasks

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

// LoadConfig recursively fetches and parses a config file, handling includes recursively.
// It returns the flattened list of tasks without executing them.
func (r *Runner) LoadConfig(ctx context.Context, url string) (TaskList, error) {
	return r.loadConfigRecursive(ctx, url, make(map[string]bool))
}

// loadConfigRecursive fetches and parses a config file, handling includes recursively.
// visited tracks URLs to prevent cycles.
func (r *Runner) loadConfigRecursive(ctx context.Context, url string, visited map[string]bool) (TaskList, error) {
	if visited[url] {
		return nil, fmt.Errorf("circular dependency detected: %s", url)
	}
	visited[url] = true

	deck.Infof("Fetching config: %s", url)
	data, err := r.fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch failed for %s: %w", url, err)
	}

	cfg, err := parseConfigData(data)
	if err != nil {
		return nil, fmt.Errorf("parse failed for %s: %w", url, err)
	}

	var allTasks TaskList

	// Process includes first (prepend logic? or just standard flattening order)
	// Usually includes are prepended or essentially "expanded in place".
	// But since `include` is a sibling of `tasks` in our struct, it's ambiguous where they go relative to `tasks`.
	// A common pattern is: Includes get processed, then local tasks.
	for _, inc := range cfg.Includes {
		absPath, err := resolvePath(url, inc)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve path %s relative to %s: %w", inc, url, err)
		}
		subTasks, err := r.loadConfigRecursive(ctx, absPath, visited)
		if err != nil {
			return nil, err
		}
		allTasks = append(allTasks, subTasks...)
	}

	allTasks = append(allTasks, cfg.Tasks...)
	return allTasks, nil
}

// resolvePath resolves a target path relative to a base path.
// Handles both HTTP URLs and local file paths.
func resolvePath(base, target string) (string, error) {
	// If target is absolute, return it
	if strings.HasPrefix(target, "http") || filepath.IsAbs(target) {
		return target, nil
	}

	// If base is HTTP
	if strings.HasPrefix(base, "http") {
		u, err := url.Parse(base)
		if err != nil {
			return "", err
		}
		rel, err := url.Parse(target)
		if err != nil {
			return "", err
		}
		return u.ResolveReference(rel).String(), nil
	}

	// Base is local filesystem
	baseDir := filepath.Dir(base)
	return filepath.Join(baseDir, target), nil
}
