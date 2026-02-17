# Templates

Glazier Go supports [Go `text/template`](https://pkg.go.dev/text/template) syntax inside YAML config files. This allows dynamic variable substitution based on the current machine's state.

## Available Variables

| Variable | Source | Example Value |
| :--- | :--- | :--- |
| `{{.Hostname}}` | `os.Hostname()` | `DESKTOP-ABC123` |
| `{{.Stage}}` | `GLAZIER_STAGE` env var | `50` |
| `{{.Timestamp}}` | Current time | `2026-02-16T17:00:00` |
| `{{.ImageID}}` | `IMAGE_ID` env var | `win11-v2` |
| `{{.Username}}` | `USERNAME` env var | `admin` |

## Usage

### Variable Substitution
```yaml
- googet.install:
    packages:
      - base-image-{{.ImageID}}
      - config-{{.Hostname}}
```

### Conditionals
```yaml
{{if .ImageID}}
- bitlocker.enable:
    mode: tpm
    backup: true
{{end}}
```

### Default Values
```yaml
- stage.set: "{{if .Stage}}{{.Stage}}{{else}}0{{end}}"
```

## Setting Variables

Template variables are populated from the system and environment variables:

```powershell
# Set the image ID before running Glazier
set IMAGE_ID=win11-v2
set GLAZIER_STAGE=50

.\glazier.exe -config_root_path .\examples\template_example.yaml
```

## How It Works

1. Glazier **fetches** the YAML config file.
2. The **template engine** processes `{{ }}` markers, substituting values from `BuildInfo`.
3. The resulting **plain YAML** is parsed and executed normally.

> [!NOTE]
> If no template markers (`{{ }}`) are present in the config, the file is passed through unchanged. There is zero overhead for non-templated configs.
