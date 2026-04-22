## ADDED Requirements

### Requirement: List VMs
The system SHALL retrieve and display all Lima VMs by executing `limactl list --format json`. Each VM entry SHALL show the VM name and its current status (Running, Stopped, or other Lima states).

#### Scenario: VMs exist
- **WHEN** the main screen loads
- **THEN** all Lima VMs are listed in the left panel with their name and status

#### Scenario: limactl list fails
- **WHEN** `limactl list` returns a non-zero exit code
- **THEN** the system displays an error message in the left panel
- **THEN** the system does not crash

---

### Requirement: VM detail view
The system SHALL display the following details for the selected VM in the right panel: name, status, CPUs, RAM (GiB), disk (GiB), and all active mounts with their read/write mode.

#### Scenario: Running VM selected
- **WHEN** a Running VM is selected
- **THEN** the right panel shows: name, status "Running", CPUs, RAM, disk, and mount list

#### Scenario: VM with no mounts
- **WHEN** a VM with no extra mounts is selected
- **THEN** the mounts section shows "No additional mounts"

---

### Requirement: Create VM
The system SHALL provide a multi-step form to create a new Lima VM. The form SHALL collect: VM name, CPUs (1 to half of host CPUs, max 4), RAM (1/2/4/8 GB), disk (20/40/60/80/100 GB), and optional host directory mounts. The system SHALL show a summary before confirming creation. On confirmation, the system SHALL execute `limactl create` with the embedded template and stream output in real time.

#### Scenario: Successful VM creation
- **WHEN** the user completes the creation form and confirms
- **THEN** the system runs `limactl create --name <name> --cpus <n> --memory <n> --disk <n> <template>` followed by `limactl start <name>`
- **THEN** output is streamed in real time in the log panel
- **THEN** on success the VM list is refreshed and the new VM is selected

#### Scenario: VM name already exists
- **WHEN** the user enters a VM name that already exists in `limactl list`
- **THEN** the form shows an inline error and does not proceed

#### Scenario: Invalid mount path
- **WHEN** the user enters a host path that does not exist
- **THEN** the form shows an inline error and does not add the mount

#### Scenario: Creation cancelled
- **WHEN** the user presses Escape at the summary step
- **THEN** no VM is created and the main screen is restored

---

### Requirement: Start VM
The system SHALL start a stopped VM by executing `limactl start <name>`. The operation SHALL stream output in the log panel.

#### Scenario: Start a stopped VM
- **WHEN** the user selects Start from the VM action menu
- **THEN** the system executes `limactl start <name>`
- **THEN** output is streamed in the log panel
- **THEN** on success the VM status updates to Running in the list

#### Scenario: Start an already running VM
- **WHEN** the user selects Start for a VM that is already Running
- **THEN** the system shows an informational message and does not execute the command

---

### Requirement: Stop VM
The system SHALL stop a running VM by executing `limactl stop <name>`. The operation SHALL complete with a status update.

#### Scenario: Stop a running VM
- **WHEN** the user selects Stop from the VM action menu
- **THEN** the system executes `limactl stop <name>`
- **THEN** on success the VM status updates to Stopped in the list

---

### Requirement: Stop all VMs
The system SHALL stop all currently running VMs when the user selects "Stop all VMs" from the global menu. The system SHALL show a confirmation dialog before proceeding.

#### Scenario: Stop all with running VMs
- **WHEN** the user selects Stop all VMs and confirms
- **THEN** the system executes `limactl stop` for each Running VM
- **THEN** all VM statuses update to Stopped

#### Scenario: Stop all with no running VMs
- **WHEN** the user selects Stop all VMs and no VMs are running
- **THEN** the system shows an informational message "No VMs are running"

---

### Requirement: Delete VM
The system SHALL delete a VM by executing `limactl stop <name>` (if running) followed by `limactl delete <name>`. The system SHALL require explicit confirmation before deleting. The confirmation dialog SHALL display the VM name.

#### Scenario: Delete a VM
- **WHEN** the user selects Delete from the VM action menu and confirms
- **THEN** the system stops the VM if running
- **THEN** the system executes `limactl delete <name>`
- **THEN** the VM is removed from the list

#### Scenario: Delete cancelled
- **WHEN** the user selects Delete but presses Escape or selects No in the confirmation dialog
- **THEN** no action is taken and the main screen is restored

---

### Requirement: Shell access
The system SHALL open an interactive shell inside a VM by suspending the TUI and executing `limactl shell <name>`. When the shell session ends, the TUI SHALL resume.

#### Scenario: Open shell in running VM
- **WHEN** the user selects Shell from the VM action menu and the VM is Running
- **THEN** the TUI suspends
- **THEN** `limactl shell <name>` runs with full terminal control
- **THEN** when the user exits the shell, the TUI resumes

#### Scenario: Open shell in stopped VM
- **WHEN** the user selects Shell from the VM action menu and the VM is Stopped
- **THEN** the system starts the VM first
- **THEN** opens the shell after the VM is running
