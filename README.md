# Glazier Go

A modern, high-performance rewrite of the [Glazier](https://github.com/google/glazier) imaging system in Go.

## 💡 Inspiration & Motivation

The original Glazier is a powerful imaging tool written in Python. While robust, deploying Python applications in a pre-boot environment (WinPE) presents challenges:
- **Dependency Hell**: Managing Python runtime, pip packages, and native extensions.
- **Size**: Python + deps can be large.
- **Performance**: Startup time and memory footprint.
- **Type Safety**: Runtime errors that could be caught at compile time.

**Glazier Go solves these problems**:
- **Single Static Binary**: `glazier.exe` (no Python installation required).
- **Type Safety**: Leveraging Go's strong typing for robust configuration parsing.
- **Concurrency**: Native Goroutines for parallel task execution.
- **Clean Architecture**: Decoupling the core logic from the Windows API wrappers.

## 🏗️ Architecture

We employ a **Plugin-style Architecture** to separate the "Business Logic" from the "System Actions":

1.  **Core Engine (`internal/config`)**:
    - Handles YAML parsing, policy validation, and flow control.
    - Pure Go logic, platform-agnostic, easily unit-tested.

2.  **Action Registry (`internal/actions`)**:
    - A dynamic registry where actions like `bitlocker.enable` or `domain.join` are registered.
    - Ensures new capabilities can be added without modifying the core engine.

3.  **Legacy Wrappers (`go/`)**:
    - We leverage the battle-tested low-level libraries from the original Google repo (e.g., `go/bitlocker`, `go/power`).
    - Our `internal/actions` wrappers adapt these APIs to a unified `Action` interface.

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- Windows (for running the binary, though it builds on Linux/macOS)

### Building
```powershell
go mod tidy
go build ./cmd/glazier
```

### Running
Glazier expects a configuration root. By default, it looks for `config.yaml` or `build.yaml` (configurable).

```powershell
# Run with local examples
.\glazier.exe -config_root_path ./examples/basic.yaml
```

## 📚 Documentation

- **[Configuration Guide](docs/configuration.md)**: Learn how to structure your YAML files.
- **[Action Reference](docs/actions.md)**: Full list of supported actions (`bitlocker`, `googet`, etc.) and their parameters.
- **[Manual Testing](MANUAL_TESTING.md)**: How to verify the binary in a real environment.

## 🔮 Next Steps

- [ ] **WinPE Integration**: Test the binary in a live WinPE boot environment.
- [ ] **UI Layer**: Add a simple UI (perhaps using [walk](https://github.com/lxn/walk)) for user prompts.
- [ ] **More Actions**: Port remaining Python actions (e.g., advanced registry tweaks, complex file operations).
- [ ] **Remote Config**: Support fetching configs from HTTP/HTTPS sources (Engine supports it, needs flags).

## 📄 License
Apache 2.0 (Same as original Glazier)
