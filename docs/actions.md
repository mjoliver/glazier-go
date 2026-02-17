# Supported Actions

This document lists all available actions in Glazier Go and their configuration parameters.

## BitLocker (`bitlocker.enable`)
Enables BitLocker encryption on the system drive.

| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `mode` | string | Yes | Encryption mode: `tpm`, `password`. |
| `backup` | bool | No | Whether to backup recovery key to AD. |

```yaml
- bitlocker.enable:
    mode: tpm
    backup: true
```

## Domain Join (`domain.join`)
Joins the machine to a Windows Domain.

| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `domain` | string | Yes | The FQDN of the domain (e.g., `example.com`). |
| `ou` | string | No | Distinguished Name of the OU. |
| `user` | string | No | Username with join privileges. |
| `password` | string | No | Password for the user. |

```yaml
- domain.join:
    domain: example.com
    user: join_user
    password: secret
```

## GooGet (`googet.install`)
Installs software packages using GooGet.

| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `packages` | list | Yes | List of package names to install. |
| `reinstall` | bool | No | Force reinstall if already present. |
| `db_only` | bool | No | Update local DB only (no actual install). |

**Shorthand Syntax:**
```yaml
- googet.install:
    - package-a
    - package-b
```

**Full Syntax:**
```yaml
- googet.install:
    packages: [package-a, package-b]
    reinstall: true
```

## Partition Disk (`partition.disk`)
Partitions a specific disk.

| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `disk_id` | int | Yes | The disk number (e.g., 0). |
| `partition_id` | int | Yes | The partition number. |
| `label` | string | No | Volume label. |
| `assign_letter` | bool | No | Whether to assign a drive letter. |

```yaml
- partition.disk:
    disk_id: 0
    partition_id: 1
    label: "Data"
```

## Power (`system.power`)
Performs power management operations.

| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `type` | string | Yes | `reboot` or `shutdown`. |
| `delay` | int | No | Delay in seconds. |
| `reason` | string | No | Reason code (`maintenance`, `installation`, `upgrade`). |
| `force` | bool | No | Force the action. |

```yaml
- system.power:
    type: reboot
    delay: 30
    reason: installation
```

## Stage (`stage.set`)
Sets the current build stage in the registry.

| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | string/int | Yes | The stage ID to set. |

```yaml
- stage.set: 10
```

## Task (`task.create`)
Creates a scheduled task.

| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `name` | string | Yes | Name of the task. |
| `command` | string | Yes | Executable to run. |
| `args` | list | No | Arguments for the command. |
| `trigger` | string | No | `boot` or `time` (default: time + 1 min). |

```yaml
- task.create:
    name: "Finalize"
    command: "C:\\Windows\\System32\\cmd.exe"
    args: ["/c", "echo done"]
    trigger: "boot"
```
