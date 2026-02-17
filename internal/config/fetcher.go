package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/glazier/internal/template"
)

// FetcherInterface allows mocking in tests.
type FetcherInterface interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
}

// Fetcher retrieves configuration files.
type Fetcher struct {
	buildInfo *template.BuildInfo
}

// NewFetcher creates a new Fetcher with optional template support.
func NewFetcher(buildInfo *template.BuildInfo) *Fetcher {
	return &Fetcher{buildInfo: buildInfo}
}

// Fetch retrieves the content at the given path/URL.
func (f *Fetcher) Fetch(ctx context.Context, path string) ([]byte, error) {
	var data []byte
	var err error

	if strings.HasPrefix(path, "http") {
		data, err = f.fetchRemote(ctx, path)
	} else {
		data, err = f.fetchLocal(path)
	}

	if err != nil {
		return nil, err
	}

	// Apply template processing if BuildInfo is available
	return template.Process(data, f.buildInfo)
}

func (f *Fetcher) fetchLocal(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f *Fetcher) fetchRemote(ctx context.Context, url string) ([]byte, error) {
	var data []byte
	var lastErr error

	// Exponential backoff configuration
	maxRetries := 3
	baseDelay := 1 * time.Second
	client := &http.Client{Timeout: 30 * time.Second}

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries-1 {
				delay := baseDelay * time.Duration(1<<uint(attempt)) // 1s, 2s, 4s
				time.Sleep(delay)
				continue
			}
			return nil, fmt.Errorf("failed to fetch URL after %d attempts: %w", maxRetries, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("bad status code: %d", resp.StatusCode)
			if attempt < maxRetries-1 {
				delay := baseDelay * time.Duration(1<<uint(attempt))
				time.Sleep(delay)
				continue
			}
			return nil, lastErr
		}

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		return data, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
