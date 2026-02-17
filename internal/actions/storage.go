package actions

import (
	"context"
	"fmt"

	"github.com/google/deck"
	"github.com/google/glazier/go/storage"
	"gopkg.in/yaml.v3"
)

// function hooks for mocking
var (
	diskClear = func(d *storage.Disk, removeData, removeOEM, zeroDisk bool) (storage.ExtendedStatus, error) {
		return d.Clear(removeData, removeOEM, zeroDisk)
	}
	getDisks = func(svc storage.Service, filter string) (storage.DiskSet, error) {
		return svc.GetDisks(filter)
	}
	storageConnect = func() (storage.Service, error) {
		return storage.Connect()
	}
	closeService = func(s *storage.Service) {
		s.Close()
	}
)

type PartitionConfig struct {
	DiskID       int    `yaml:"disk_id"`
	PartitionID  int    `yaml:"partition_id"`
	Label        string `yaml:"label"`
	AssignLetter bool   `yaml:"assign_letter"`
}

type DiskWipeConfig struct {
	DiskID     int  `yaml:"disk_id"`
	RemoveData bool `yaml:"remove_data"`
	RemoveOEM  bool `yaml:"remove_oem"`
	ZeroDisk   bool `yaml:"zero_disk"`
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

func NewDiskWipe(ctx context.Context, yamlData interface{}) (Action, error) {
	var cfg DiskWipeConfig
	// Set defaults: remove_data=true, remove_oem=true, zero_disk=false
	cfg.RemoveData = true
	cfg.RemoveOEM = true

	data, err := yaml.Marshal(yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &DiskWipe{Config: cfg}, nil
}

type Partition struct {
	Config PartitionConfig
}

func (a *Partition) Run(ctx context.Context) error {
	deck.Infof("Partitioning disk %d, part %d (label: %s)", a.Config.DiskID, a.Config.PartitionID, a.Config.Label)

	// Connect to Storage Service (WMI)
	svc, err := storage.Connect()
	if err != nil {
		return fmt.Errorf("storage.Connect: %w", err)
	}
	defer svc.Close()

	// Find the specific disk
	filter := fmt.Sprintf("WHERE Number=%d", a.Config.DiskID)
	diskSet, err := svc.GetDisks(filter)
	if err != nil {
		return fmt.Errorf("GetDisks(%s): %w", filter, err)
	}
	defer diskSet.Close()

	if len(diskSet.Disks) == 0 {
		return fmt.Errorf("disk %d not found", a.Config.DiskID)
	}
	disk := diskSet.Disks[0]

	// Create Partition
	// method signature: CreatePartition(size, useMax, offset, align, letter, assignLetter, mbr, gpt, hidden, active)
	// We default to: Max Size, No Logic for MBR/GPT specific types yet (nil), Not Hidden, Not Active
	part, _, err := disk.CreatePartition(0, true, 0, 0, "", false, nil, nil, false, false)
	if err != nil {
		return fmt.Errorf("CreatePartition (disk %d): %w", a.Config.DiskID, err)
	}
	defer part.Close()

	// Format if Label is provided
	if a.Config.Label != "" {
		// formatting requires a Volume object, which is separate from Partition in WMI.
		// To do this, we'd need to:
		// 1. Assign a temporary path/letter if none exists
		// 2. Query MSFT_Volume where Path matches
		// 3. Call Format on the Volume
		// This is complex to do blindly w/o knowing the drive letter strategy.
		// For now, we log a warning.
		deck.Warningf("Formatting is not yet implemented (label: %s). Please format manually or via script.", a.Config.Label)
	}

	if a.Config.AssignLetter {
		// AddAccessPath("", true) assigns the next available drive letter
		if _, err := part.AddAccessPath("", true); err != nil {
			return fmt.Errorf("failed to assign drive letter: %w", err)
		}
		deck.Infof("Drive letter assigned to partition %d", a.Config.PartitionID)
	}

	return nil
}

func (a *Partition) Validate() error {
	return nil
}

func init() {
	Register("partition.disk", NewPartition)
	Register("disk.wipe", NewDiskWipe)
}

type DiskWipe struct {
	Config DiskWipeConfig
}

func (a *DiskWipe) Run(ctx context.Context) error {
	deck.Infof("Wiping disk %d (data: %v, oem: %v, zero: %v)", a.Config.DiskID, a.Config.RemoveData, a.Config.RemoveOEM, a.Config.ZeroDisk)

	svc, err := storageConnect()
	if err != nil {
		return fmt.Errorf("storage.Connect: %w", err)
	}
	defer closeService(&svc)

	// Find the specific disk
	filter := fmt.Sprintf("WHERE Number=%d", a.Config.DiskID)
	diskSet, err := getDisks(svc, filter)
	if err != nil {
		return fmt.Errorf("GetDisks(%s): %w", filter, err)
	}
	defer diskSet.Close()

	if len(diskSet.Disks) == 0 {
		return fmt.Errorf("disk %d not found", a.Config.DiskID)
	}
	disk := diskSet.Disks[0]

	// Perform Wipe using the mockable function
	deck.Warningf("Executing Disk Wipe on Disk %d. This is destructive.", a.Config.DiskID)
	_, err = diskClear(&disk, a.Config.RemoveData, a.Config.RemoveOEM, a.Config.ZeroDisk)
	if err != nil {
		return fmt.Errorf("Disk.Clear failed: %w", err)
	}

	deck.Infof("Disk %d wiped successfully", a.Config.DiskID)
	return nil
}

func (a *DiskWipe) Validate() error {
	if a.Config.DiskID < 0 {
		return fmt.Errorf("disk_id must be non-negative")
	}
	return nil
}
