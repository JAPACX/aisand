## ADDED Requirements

### Requirement: Install opencode bundle
The system SHALL install the opencode tool bundle inside a VM by piping the embedded `opencode.sh` script to `limactl shell <name> bash -s`. The operation SHALL stream stdout and stderr in real time in the log panel. The script SHALL be idempotent — safe to run multiple times.

#### Scenario: Install opencode on running VM
- **WHEN** the user selects "Install tool" from the VM action menu and the VM is Running
- **THEN** the system pipes the embedded `opencode.sh` to `limactl shell <name> bash -s`
- **THEN** all output (stdout + stderr) is streamed line by line in the log panel
- **THEN** on success the log panel shows the final status summary from the script

#### Scenario: Install opencode on stopped VM
- **WHEN** the user selects "Install tool" and the VM is Stopped
- **THEN** the system starts the VM first
- **THEN** proceeds with the installation

#### Scenario: Script already installed (idempotent)
- **WHEN** the user runs Install tool on a VM where opencode is already installed
- **THEN** the script reports each component as already installed
- **THEN** no reinstallation occurs
- **THEN** the log panel shows `[✓]` status for each component

#### Scenario: Installation fails
- **WHEN** the script exits with a non-zero code
- **THEN** the log panel shows the error output
- **THEN** the system displays "Installation failed. Press any key to return."

---

### Requirement: Real-time log panel
The system SHALL display a scrollable log panel for any long-running operation. The panel SHALL append new lines as they arrive from the subprocess stdout/stderr. The panel SHALL automatically scroll to the latest line. When the operation completes, the system SHALL display the exit status and prompt the user to press any key to return to the main screen.

#### Scenario: Log panel during operation
- **WHEN** a long-running operation is in progress
- **THEN** each output line appears in the log panel as it is received
- **THEN** the panel scrolls to show the latest line

#### Scenario: Operation completes successfully
- **WHEN** the subprocess exits with code 0
- **THEN** the log panel shows "✓ Done. Press any key to continue."

#### Scenario: Operation completes with error
- **WHEN** the subprocess exits with a non-zero code
- **THEN** the log panel shows "✗ Failed (exit <code>). Press any key to continue."
