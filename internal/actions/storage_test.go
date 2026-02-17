package actions

import (
	"context"
	"testing"
)

func TestPartition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  PartitionConfig
		wantErr bool
	}{
		{"valid", PartitionConfig{DiskID: 0, PartitionID: 1}, false},
		// In the current implementation, Validate() always returns nil.
		// If we add validation logic later (e.g., DiskID >= 0), add tests here.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Partition{Config: tt.config}
			if err := a.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewPartition(t *testing.T) {
	ctx := context.Background()
	yamlData := map[string]interface{}{
		"disk_id":       0,
		"partition_id":  1,
		"label":         "Data",
		"assign_letter": true,
	}

	a, err := NewPartition(ctx, yamlData)
	if err != nil {
		t.Fatalf("NewPartition() error = %v", err)
	}

	p, ok := a.(*Partition)
	if !ok {
		t.Fatalf("NewPartition() returned wrong type")
	}

	if p.Config.DiskID != 0 {
		t.Errorf("DiskID = %d, want 0", p.Config.DiskID)
	}
	if p.Config.PartitionID != 1 {
		t.Errorf("PartitionID = %d, want 1", p.Config.PartitionID)
	}
	if p.Config.Label != "Data" {
		t.Errorf("Label = %q, want Data", p.Config.Label)
	}
	if !p.Config.AssignLetter {
		t.Errorf("AssignLetter = false, want true")
	}
}
