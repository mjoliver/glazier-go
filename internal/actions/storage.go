package actions

import (
	"context"
	"fmt"

	"github.com/google/deck"
	// "github.com/google/glazier/go/storage"
	"gopkg.in/yaml.v3"
)

type PartitionConfig struct {
	DiskID       int    `yaml:"disk_id"`
	PartitionID  int    `yaml:"partition_id"`
	Label        string `yaml:"label"`
	AssignLetter bool   `yaml:"assign_letter"`
}

func NewPartition(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg PartitionConfig
	data, err := yaml.Marshal(yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &Partition{Config: cfg}, nil
}

type Partition struct {
	Config PartitionConfig
}

func (a *Partition) Run(ctx context.Context) error {
	deck.Infof("Partitioning disk %d, part %d", a.Config.DiskID, a.Config.PartitionID)

	// Example call to existing library
	// d := storage.Disk{Number: a.Config.DiskID}
	// err := d.Partition(a.Config.PartitionID, ...)
	// logic would go here

	return nil
}

func (a *Partition) Validate() error {
	return nil
}

func init() {
	Register("partition.disk", NewPartition)
}
