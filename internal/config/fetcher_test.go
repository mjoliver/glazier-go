package config

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetcher_Fetch(t *testing.T) {
	// Test HTTP path handling
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer server.Close()

	f := NewFetcher(nil)
	data, err := f.Fetch(context.Background(), server.URL)

	if err != nil {
		t.Errorf("Fetch() unexpected error = %v", err)
	}

	if string(data) != "test" {
		t.Errorf("Fetch() got %q, want %q", string(data), "test")
	}
}

func TestFetcher_fetchRemote_Success(t *testing.T) {
	// Create a test server that returns success
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test content"))
	}))
	defer server.Close()

	f := NewFetcher(nil)
	data, err := f.fetchRemote(context.Background(), server.URL)

	if err != nil {
		t.Fatalf("fetchRemote() unexpected error = %v", err)
	}

	if string(data) != "test content" {
		t.Errorf("fetchRemote() got data = %q, want %q", string(data), "test content")
	}
}

func TestFetcher_fetchRemote_Retry(t *testing.T) {
	attempts := 0
	maxAttempts := 3

	// Create a test server that fails twice, then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		t.Logf("HTTP attempt %d/%d", attempts, maxAttempts)
		if attempts < maxAttempts {
			w.WriteHeader(http.StatusInternalServerError)
			t.Logf("  → Returning 500 error (will retry)")
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success after retries"))
		t.Logf("  → Returning 200 OK")
	}))
	defer server.Close()

	f := NewFetcher(nil)
	start := time.Now()
	data, err := f.fetchRemote(context.Background(), server.URL)
	elapsed := time.Since(start)

	t.Logf("Total attempts: %d, elapsed time: %v", attempts, elapsed)

	if err != nil {
		t.Fatalf("fetchRemote() unexpected error = %v", err)
	}

	if string(data) != "success after retries" {
		t.Errorf("fetchRemote() got data = %q, want %q", string(data), "success after retries")
	}

	if attempts != maxAttempts {
		t.Errorf("fetchRemote() made %d attempts, want %d", attempts, maxAttempts)
	}

	// Check that retries took some time (exponential backoff: ~3s for 2 retries)
	minExpectedTime := 2 * time.Second
	if elapsed < minExpectedTime {
		t.Errorf("fetchRemote() took %v, expected at least %v for retries", elapsed, minExpectedTime)
	} else {
		t.Logf("✓ Exponential backoff verified: %v", elapsed)
	}
}

func TestFetcher_fetchRemote_MaxRetriesExceeded(t *testing.T) {
	attempts := 0

	// Create a test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	f := NewFetcher(nil)
	_, err := f.fetchRemote(context.Background(), server.URL)

	if err == nil {
		t.Fatal("fetchRemote() expected error after max retries, got nil")
	}

	if attempts != 3 {
		t.Errorf("fetchRemote() made %d attempts, want 3", attempts)
	}

	if !contains(err.Error(), "bad status code") {
		t.Errorf("fetchRemote() error = %v, want error containing 'bad status code'", err)
	}
}

func TestFetcher_fetchRemote_ContextCanceled(t *testing.T) {
	// Create a test server with a delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	f := NewFetcher(nil)
	_, err := f.fetchRemote(ctx, server.URL)

	if err == nil {
		t.Fatal("fetchRemote() expected error from context cancellation, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) && !contains(err.Error(), "context") {
		t.Errorf("fetchRemote() error = %v, want context cancellation error", err)
	}
}
