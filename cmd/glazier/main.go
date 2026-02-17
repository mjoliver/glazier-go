package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/mjoliver/glazier-go/internal/config"
	"github.com/mjoliver/glazier-go/internal/template"
)

var (
	configRootPath = flag.String("config_root_path", "/", "Root path to configuration files")
	ntpServer      = flag.String("ntp_server", "time.google.com", "NTP server to use for time synchronization")
	preserveTasks  = flag.Bool("preserve_tasks", false, "Preserve the local task list on startup")
	verifyUrls     = flag.String("verify_urls", "", "Comma-separated list of URLs to verify reachability")
)

func main() {
	flag.Parse()

	// Initialize logging (using standard log for now, could switch to deck later)
	log.SetOutput(os.Stdout)
	log.Println("Starting Glazier Go...")

	// Todo: Initialize BuildInfo
	// Todo: Check WinPE status
	// Todo: Sync Clock
	// Todo: Check Battery

	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatalf("Glazier failed: %v", err)
	}
}

func run(ctx context.Context) error {
	log.Printf("Config Root Path: %s", *configRootPath)

	// Initialize build info for templates
	buildInfo, err := template.NewBuildInfo()
	if err != nil {
		log.Printf("Warning: failed to initialize build info: %v", err)
		buildInfo = nil // Continue without template support
	} else {
		log.Printf("Build Info: Hostname=%s, Stage=%s", buildInfo.Hostname, buildInfo.Stage)
	}

	// Create Config Runner
	fetcher := config.NewFetcher(buildInfo)
	runner := config.NewRunner(fetcher)

	// Execute
	return runner.Start(ctx, *configRootPath)
}
