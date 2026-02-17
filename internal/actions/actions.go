package actions

import (
	"context"
	"fmt"
)

// Action represents a single step in the Glazier imaging process.
type Action interface {
	// Run execute the action.
	Run(ctx context.Context) error

	// Validate checks if the action configuration is valid.
	Validate() error
}

// Factory functions create new Actions from raw YAML data (usually map[string]interface{}).
type Factory func(ctx context.Context, yamlData interface{}) (Action, error)

var Registry = map[string]Factory{}

// Register adds an action factory to the global registry.
func Register(name string, factory Factory) {
	if _, exists := Registry[name]; exists {
		panic(fmt.Sprintf("Action %s already registered", name))
	}
	Registry[name] = factory
}

// New creates an action instance by name.
func New(ctx context.Context, name string, yamlData interface{}) (Action, error) {
	factory, ok := Registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", name)
	}
	return factory(ctx, yamlData)
}
