package actions

import (
	"context"
	"fmt"

	"github.com/google/deck"
	"github.com/google/glazier/go/stages"
	"gopkg.in/yaml.v3"
)

type StageSet struct {
	ID string `yaml:"id"`
}

func NewStageSet(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg StageSet
	// Handle single string case: "stage.set: 10"
	if id, ok := yamlData.(string); ok {
		cfg.ID = id
	} else if id, ok := yamlData.(int); ok {
		cfg.ID = fmt.Sprintf("%d", id)
	} else {
		data, err := yaml.Marshal(yamlData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	return &StageSet{ID: cfg.ID}, nil
}

func (a *StageSet) Run(ctx context.Context) error {
	deck.Infof("Setting stage to: %s", a.ID)
	// Call existing library
	return stages.SetStage(a.ID, stages.StartKey)
}

func (a *StageSet) Validate() error {
	if a.ID == "" {
		return fmt.Errorf("stage ID is required")
	}
	return nil
}

func init() {
	Register("stage.set", NewStageSet)
}
