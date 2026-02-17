# Glazier Go - Manual Integration Test Plan

Since many Glazier actions (BitLocker, Domain Join, Power) interact directly with the Windows OS, automated unit tests can only verify the *logic*, not the *effect*. This plan outlines the manual steps to verify the binary in a real environment (VM or Test Device).

## Prerequisites
- A Windows 10/11 Virtual Machine (or physical test device).
- `glazier.exe` (built from `cmd/glazier`).
- `examples/` folder (containing `basic.yaml` and `full.yaml`).

## Test Cases

### 1. Basic Policy & Config Flow
**Goal**: Verify the engine runs, checks policy, and executes simple actions.
**Config**: `examples/basic.yaml`

**Steps:**
1.  Copy `glazier.exe` and `examples/` to `C:\Test`.
2.  Open PowerShell as Administrator.
3.  Run: `.\glazier.exe -config_root_path .\examples\basic.yaml`

**Expected Outcome:**
- [ ] Logs show "Starting Glazier Go..."
- [ ] Logs show "Checking policy: os_version" -> "Passed".
- [ ] Logs show "Executing action: stage.set" -> "10".
- [ ] Logs show "Executing action: googet.install".
- [ ] **System Reboots** after 30 seconds.

### 2. Full Imaging Workflow (Destructive)
**Goal**: Verify complex actions including Encryption and Domain Join.
**Config**: `examples/full.yaml`
**⚠️ WARNING**: This will attempt to encrypt the drive and join a domain. Run only on a disposable VM.

**Steps:**
1.  Edit `examples/full.yaml` to set valid Domain credentials if testing join, or expect failure.
2.  Run: `.\glazier.exe -config_root_path .\examples\full.yaml`

**Expected Outcome:**
- [ ] **BitLocker**: Methods called (check logs). If TPM is present, encryption starts.
- [ ] **GooGet**: Mock packages processed.
- [ ] **Domain Join**: Attempts to contact DC (check logs for success/failure message).
- [ ] **Scheduled Task**: Run `taskschd.msc` and verify "FinalizeSetup" task exists.
- [ ] **Power**: System reboots.

### 3. Failure Handling
**Goal**: Verify Glazier stops on errors.
**Config**: Create `error.yaml`

```yaml
- googet.install:
    packages: ["non-existent-package"]
- system.power:
    type: reboot
```

**Expected Outcome:**
- [ ] Logs show error during GooGet install.
- [ ] **System DOES NOT reboot** (execution halts).

## Troubleshooting
- **Logs**: Check console output. Redirect to file if needed: `.\glazier.exe ... > log.txt 2>&1`
- **Permissions**: Ensure you are running as Administrator.
