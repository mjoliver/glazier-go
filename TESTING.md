# Testing Guide

## Running Tests

### Basic Commands

```bash
# Run all tests (brief output)
go test ./internal/...

# Run with verbose output (shows all t.Logf() messages)
go test ./internal/... -v

# Run specific package
go test ./internal/config -v

# Run specific test
go test ./internal/config -v -run TestFetcher_fetchRemote_Retry

# Run tests matching a pattern
go test ./internal/... -v -run ".*Validate"
```

### Enhanced Output

```bash
# Show test coverage
go test ./internal/... -cover

# Detailed coverage report
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# JSON output (for CI/CD)
go test ./internal/... -json

# Show test timing
go test ./internal/... -v -bench=.
```

### Debugging

```bash
# Run with race detector (catches concurrency bugs)
go test ./internal/... -race

# Run tests multiple times (catch flaky tests)
go test ./internal/config -v -count=5

# Fail fast (stop on first failure)
go test ./internal/... -v -failfast
```

## Test Output Example

With `-v` flag, you'll see:
```
=== RUN   TestFetcher_fetchRemote_Retry
    fetcher_test.go:59: HTTP attempt 1/3
    fetcher_test.go:62:   → Returning 500 error (will retry)
    fetcher_test.go:59: HTTP attempt 2/3
    fetcher_test.go:62:   → Returning 500 error (will retry)
    fetcher_test.go:59: HTTP attempt 3/3
    fetcher_test.go:67:   → Returning 200 OK
    fetcher_test.go:76: Total attempts: 3, elapsed time: 3.002s
    fetcher_test.go:95: ✓ Exponential backoff verified: 3.002s
--- PASS: TestFetcher_fetchRemote_Retry (3.00s)
```

## Platform Notes

- Tests run on **any OS** (Windows, Linux, macOS)
- Actions like BitLocker/Domain are **Windows-specific** when executed
- Test logic validates YAML parsing, validation, and control flow (platform-agnostic)
