# Configuration Structure

Glazier uses a YAML-based configuration system. The configuration is a **list of tasks** executed sequentially.

## File Format

The root configuration file is a YAML list. Each item in the list is a map containing either a **Policy Check** or an **Action**.

```yaml
# Example Configuration
- policy:
  - os_version

- action_name:
    param1: value
    param2: value
```

## Control Flow

1.  **Sequential Execution**: Tasks are executed in the order they appear in the file.
2.  **Failure handling**: If an action fails (returns an error), execution **stops** immediately (unless specific error handling is implemented in the future).
3.  **Policy Checks**: Policies act as gates. If a policy check fails, the execution stops.

## Example

```yaml
# 1. Check if we are compliant
- policy:
    - os_version
    - device_model

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
```
