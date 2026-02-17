# Configuration Structure

Glazier uses a YAML-based configuration system. The configuration is a **list of tasks** executed sequentially.

## File Format

The root configuration file is a YAML list. Each item in the list is a map containing either a **Policy Check** or an **Action**.

```yaml
# Example Configuration
- policy:
  - os_version:
      version: "11"
  - device_model:
      allowed: ["Nitro", "ThinkPad"]

- action_name:
    param1: value
    param2: value
```

## Control Flow

1.  **Sequential Execution**: Tasks are executed in the order they appear in the file.
2.  **Failure handling**: If an action fails (returns an error), execution **stops** immediately (unless specific error handling is implemented in the future).
3.  **Policy Checks**: Policies act as gates. If a policy check fails, the execution stops. Version checks are **exact match** — a config locked to `"Server 2019"` will not run on a `"Server 2022"` host.

## Example

```yaml
# 1. Check if we are compliant
- policy:
    - os_version:
        version: "11"
    - device_model:
        allowed: ["Nitro"]
    - chassis_type:
        allowed: ["laptop", "desktop"]

# 2. Set Build Stage
- stage.set: "10"

# 3. Perform Action
- googet.install:
    packages:
      - google-chrome-stable
```

## Running Glazier

To run Glazier with a specific config file:

```powershell
.\glazier.exe -config_root_path path/to/config.yaml
```

## Templates

Config files support Go `text/template` syntax for dynamic values. See [Templates Reference](templates.md) for full details.

```yaml
- googet.install:
    packages:
      - base-{{.Hostname}}
      - image-{{.ImageID}}

## Error Handling & Retries

Any action can be configured with automatic retries and error handling behavior.

### `retries` (int)
Number of times to retry a failed action. Retries use exponential backoff (1s, 2s, 4s...). Default is `0`.

### `on_error` (string)
Behavior when an action fails (after all retries).
- `fail` (default): Stop execution and exit with error.
- `continue`: Log a warning and proceed to the next task.

```yaml
- file.download:
    url: https://example.com/installer.exe
    dst: C:\installer.exe
    retries: 3           # Retry up to 3 times (total 4 attempts)
    on_error: continue   # If it still fails, log warning and continue
```
```
