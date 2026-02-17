package template

import (
	"fmt"
	"os"
	"time"
)

// BuildInfo holds template variables for config processing.
type BuildInfo struct {
	Hostname  string
	Stage     string
	Timestamp string
	ImageID   string
	Username  string
}

// NewBuildInfo collects system information for template context.
func NewBuildInfo() (*BuildInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get stage from env var (set by stage.set action or externally)
	stage := os.Getenv("GLAZIER_STAGE")
	if stage == "" {
		stage = "0"
	}

	username := os.Getenv("USERNAME")
	if username == "" {
		username = "SYSTEM"
	}

	return &BuildInfo{
		Hostname:  hostname,
		Stage:     stage,
		Timestamp: time.Now().Format("2006-01-02T15:04:05"),
		ImageID:   os.Getenv("IMAGE_ID"),
		Username:  username,
	}, nil
}

// Get returns the value of a template variable by name.
func (b *BuildInfo) Get(key string) (string, error) {
	switch key {
	case "Hostname":
		return b.Hostname, nil
	case "Stage":
		return b.Stage, nil
	case "Timestamp":
		return b.Timestamp, nil
	case "ImageID":
		return b.ImageID, nil
	case "Username":
		return b.Username, nil
	default:
		return "", fmt.Errorf("unknown template variable: %s", key)
	}
}
