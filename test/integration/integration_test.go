//go:build windows

package integration

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	glazierReg "github.com/google/glazier/go/registry"
	"github.com/mjoliver/glazier-go/internal/config"
)

// TestIntegration_EndToEnd parses and executes a real config, then verifies
// that registry values were set and files were created on the Windows runner.
func TestIntegration_EndToEnd(t *testing.T) {
	// Find the project root (test/integration -> project root)
	_, thisFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

	configPath := filepath.Join(projectRoot, "test", "integration", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config not found at %s: %v", configPath, err)
	}

	// Clean up before and after
	cleanup := func() {
		os.RemoveAll(`C:\GlazierTest`)
		glazierReg.Delete(`SOFTWARE\GlazierTest`, "BuildVersion")
		glazierReg.Delete(`SOFTWARE\GlazierTest`, "BuildStage")
	}
	cleanup()
	t.Cleanup(cleanup)

	// Run the config
	ctx := context.Background()
	fetcher := config.NewFetcher(nil)
	runner := config.NewRunner(fetcher)

	if err := runner.Start(ctx, configPath); err != nil {
		t.Fatalf("Runner.Start() failed: %v", err)
	}

	// --- Verify directory was created ---
	t.Run("verify_mkdir", func(t *testing.T) {
		info, err := os.Stat(`C:\GlazierTest\integration`)
		if err != nil {
			t.Fatalf("directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Fatal("expected a directory")
		}
		t.Log("✓ C:\\GlazierTest\\integration exists")
	})

	// --- Verify registry string value ---
	t.Run("verify_registry_string", func(t *testing.T) {
		val, err := glazierReg.GetString(`SOFTWARE\GlazierTest`, "BuildVersion")
		if err != nil {
			t.Fatalf("registry string not set: %v", err)
		}
		if val != "1.0.0-ci" {
			t.Fatalf("expected '1.0.0-ci', got %q", val)
		}
		t.Logf("✓ BuildVersion = %q", val)
	})

	// --- Verify registry dword value ---
	t.Run("verify_registry_dword", func(t *testing.T) {
		val, err := glazierReg.GetInteger(`SOFTWARE\GlazierTest`, "BuildStage")
		if err != nil {
			t.Fatalf("registry dword not set: %v", err)
		}
		if val != 42 {
			t.Fatalf("expected 42, got %d", val)
		}
		t.Logf("✓ BuildStage = %d", val)
	})

	// --- Verify downloaded file ---
	t.Run("verify_download", func(t *testing.T) {
		info, err := os.Stat(`C:\GlazierTest\integration\downloaded.txt`)
		if err != nil {
			t.Fatalf("downloaded file not found: %v", err)
		}
		if info.Size() == 0 {
			t.Fatal("downloaded file is empty")
		}
		t.Logf("✓ downloaded.txt exists (%d bytes)", info.Size())
	})
}
