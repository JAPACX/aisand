## ADDED Requirements

### Requirement: List mounts
The system SHALL display all active mounts for a VM, including the mount path and read/write mode (read-only or read-write), when the user selects "Mounts" from the VM action menu.

#### Scenario: VM has mounts
- **WHEN** the user selects Mounts from the VM action menu
- **THEN** the system displays each mount with its path and mode (read-only / read-write)

#### Scenario: VM has no extra mounts
- **WHEN** the user selects Mounts and the VM has no additional mounts beyond the default home mount
- **THEN** the system displays "No additional mounts configured"

---

### Requirement: Add mount
The system SHALL allow the user to add a host directory as a read-write mount to a VM by selecting "Mount" from the VM action menu. The system SHALL validate that the path exists on the host. If the VM is running, the system SHALL warn that a restart is required and SHALL handle the stop/edit/start cycle automatically after confirmation.

#### Scenario: Add mount to stopped VM
- **WHEN** the user selects Mount, enters a valid host path, and confirms
- **THEN** the system executes `limactl edit <name> --mount <path>:w --tty=false`
- **THEN** the mount appears in the VM detail panel

#### Scenario: Add mount to running VM
- **WHEN** the user selects Mount for a running VM, enters a valid host path, and confirms restart
- **THEN** the system stops the VM
- **THEN** applies the mount via `limactl edit`
- **THEN** starts the VM again
- **THEN** the mount appears in the VM detail panel

#### Scenario: Path does not exist
- **WHEN** the user enters a host path that does not exist
- **THEN** the system shows an inline error "Path does not exist on host"
- **THEN** the mount is not added

#### Scenario: Path already mounted
- **WHEN** the user enters a host path that is already mounted in the VM
- **THEN** the system shows an inline error "Path is already mounted"
- **THEN** no duplicate mount is added

---

### Requirement: Remove mount
The system SHALL allow the user to remove a mount from a VM by selecting "Unmount" from the VM action menu. The system SHALL display a numbered list of current mounts and allow the user to select one for removal. If the VM is running, the system SHALL warn that a restart is required and SHALL handle the stop/edit/start cycle automatically after confirmation.

#### Scenario: Remove mount from stopped VM
- **WHEN** the user selects Unmount, picks a mount from the list, and confirms
- **THEN** the system removes the selected mount via `limactl edit`
- **THEN** the mount no longer appears in the VM detail panel

#### Scenario: Remove mount from running VM
- **WHEN** the user selects Unmount for a running VM, picks a mount, and confirms restart
- **THEN** the system stops the VM
- **THEN** removes the mount via `limactl edit`
- **THEN** starts the VM again

#### Scenario: Remove last mount
- **WHEN** the user removes the only non-default mount
- **THEN** the system executes `limactl edit <name> --mount-none --tty=false`
- **THEN** the mounts section shows "No additional mounts"

#### Scenario: Unmount cancelled
- **WHEN** the user selects Unmount but presses Escape before confirming
- **THEN** no changes are made
