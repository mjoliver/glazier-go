package actions

import (
	"context"
	"testing"

	"github.com/google/glazier/go/storage"
)

func TestDiskWipe_Run(t *testing.T) {
	// Save originals to restore after test
	origConnect := storageConnect
	origGetDisks := getDisks
	origDiskClear := diskClear
	origCloseService := closeService
	defer func() {
		storageConnect = origConnect
		getDisks = origGetDisks
		diskClear = origDiskClear
		closeService = origCloseService
	}()

	// Mock Connect
	storageConnect = func() (storage.Service, error) {
		return storage.Service{}, nil
	}

	// Mock Close
	closeService = func(s *storage.Service) {}

	// Mock GetDisks
	getDisks = func(svc storage.Service, filter string) (storage.DiskSet, error) {
		// Return a fake disk
		return storage.DiskSet{
			Disks: []storage.Disk{
				{Number: 0},
			},
		}, nil
	}

	// Mock Clear and capture calls
	var clearCalled bool
	var capturedRemoveData, capturedRemoveOEM, capturedZeroDisk bool
	diskClear = func(d *storage.Disk, removeData, removeOEM, zeroDisk bool) (storage.ExtendedStatus, error) {
		clearCalled = true
		capturedRemoveData = removeData
		capturedRemoveOEM = removeOEM
		capturedZeroDisk = zeroDisk
		return storage.ExtendedStatus{}, nil
	}

	// Test Case
	ctx := context.Background()
	wipeAction, err := NewDiskWipe(ctx, map[string]interface{}{
		"disk_id":     0,
		"remove_data": true,
		"remove_oem":  true,
		"zero_disk":   false,
	})
	if err != nil {
		t.Fatalf("NewDiskWipe failed: %v", err)
	}

	if err := wipeAction.Run(ctx); err != nil {
		t.Errorf("Run() failed: %v", err)
	}

	if !clearCalled {
		t.Error("Disk.Clear was not called")
	}
	if !capturedRemoveData {
		t.Error("remove_data should be true")
	}
	if !capturedRemoveOEM {
		t.Error("remove_oem should be true")
	}
	if capturedZeroDisk {
		t.Error("zero_disk should be false")
	}
}
