package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/deck"
	"github.com/google/deck/backends/logger"
	"github.com/mjoliver/glazier-go/internal/config"
	"github.com/mjoliver/glazier-go/internal/template"
)

var (
	configRootPath = flag.String("config_root_path", "/", "Root path to configuration files")
	ntpServer      = flag.String("ntp_server", "time.google.com", "NTP server to use for time synchronization")
	preserveTasks  = flag.Bool("preserve_tasks", false, "Preserve the local task list on startup")
	verifyUrls     = flag.String("verify_urls", "", "Comma-separated list of URLs to verify reachability")
	validate       = flag.Bool("validate", false, "Validate the configuration without executing (dry-run)")
)

func main() {
	flag.Parse()

	// Initialize deck logging with stdout backend
	deck.Add(logger.Init(os.Stdout, 0))
	defer deck.Close()

	if *validate {
		deck.Info("Running in VALIDATION mode")
	} else {
		deck.Info("Starting Glazier Go...")
	}

	// Todo: Initialize BuildInfo
	// Todo: Check WinPE status
	// Todo: Sync Clock
	// Todo: Check Battery

	ctx := context.Background()
	if err := run(ctx); err != nil {
		deck.Errorf("Glazier failed: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	deck.Infof("Config Root Path: %s", *configRootPath)

	// Initialize build info for templates
	buildInfo, err := template.NewBuildInfo()
	if err != nil {
		deck.Warningf("Failed to initialize build info: %v", err)
		buildInfo = nil // Continue without template support
	} else {
		deck.Infof("Build Info: Hostname=%s, Stage=%s", buildInfo.Hostname, buildInfo.Stage)
	}

	// Create Config Runner
	fetcher := config.NewFetcher(buildInfo)
	runner := config.NewRunner(fetcher)

	// Load Config
	if *validate {
		tasks, err := runner.LoadConfig(ctx, *configRootPath)
		if err != nil {
			return fmt.Errorf("config load failed: %w", err)
		}
		if err := config.Validate(ctx, tasks); err != nil {
			return err
		}
		deck.Info("Validation successful!")
		return nil
	}

	// Execute
	return runner.Start(ctx, *configRootPath)
}
