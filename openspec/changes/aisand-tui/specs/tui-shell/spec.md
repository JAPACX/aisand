## ADDED Requirements

### Requirement: Binary startup check
The system SHALL verify that `limactl` is available in PATH before rendering the main interface. If `limactl` is not found, the system SHALL display a blocking onboarding screen and SHALL NOT allow access to any VM management functionality until the dependency is resolved.

#### Scenario: limactl not installed
- **WHEN** the user runs `aisand` and `limactl` is not found in PATH
- **THEN** the system displays the Host Setup Required screen
- **THEN** all VM management options are inaccessible

#### Scenario: limactl installed
- **WHEN** the user runs `aisand` and `limactl` is found in PATH
- **THEN** the system displays the main split-panel screen directly

---

### Requirement: Split-panel main layout
The system SHALL render a split-panel layout with a left panel showing the VM list and a right panel showing the detail of the currently selected VM. Both panels SHALL be visible simultaneously without scrolling.

#### Scenario: VM selected in list
- **WHEN** the user navigates to a VM in the left panel
- **THEN** the right panel updates immediately to show that VM's details (name, status, CPUs, RAM, disk, mounts)

#### Scenario: No VMs exist
- **WHEN** there are no Lima VMs
- **THEN** the left panel shows an empty state message
- **THEN** the right panel shows a prompt to create a new VM

---

### Requirement: VM list navigation
The system SHALL allow the user to navigate the VM list using arrow keys (up/down). The currently selected VM SHALL be visually highlighted.

#### Scenario: Navigate list
- **WHEN** the user presses the down arrow key
- **THEN** the selection moves to the next VM in the list
- **THEN** the right panel updates to show the newly selected VM's details

---

### Requirement: Contextual action menu
The system SHALL display a contextual action menu when the user presses Enter on a selected VM. The menu SHALL list all available actions for that VM. Pressing Escape SHALL close the menu and return to the main screen.

#### Scenario: Open action menu
- **WHEN** the user presses Enter on a VM in the list
- **THEN** the action menu appears showing: Shell, Start, Stop, Delete, Install tool, Mount, Unmount, Mounts

#### Scenario: Close action menu
- **WHEN** the action menu is open and the user presses Escape
- **THEN** the action menu closes and the main screen is restored

---

### Requirement: Global menu
The system SHALL display a global menu when the user presses `g` from the main screen. The global menu SHALL provide: New VM, Stop all VMs, Host setup, Refresh. Pressing Escape SHALL close the menu.

#### Scenario: Open global menu
- **WHEN** the user presses `g` on the main screen
- **THEN** the global menu appears with options: New VM, Stop all VMs, Host setup, Refresh

#### Scenario: Refresh from global menu
- **WHEN** the user selects Refresh from the global menu
- **THEN** the VM list is reloaded from `limactl list`
- **THEN** the main screen is restored with updated data

---

### Requirement: Status bar with keybindings
The system SHALL display a persistent status bar at the bottom of the screen showing the available keybindings for the current screen context.

#### Scenario: Main screen keybindings
- **WHEN** the main screen is active
- **THEN** the status bar shows: `↑↓ navigate  enter actions  g global  q quit`

---

### Requirement: Quit
The system SHALL exit cleanly when the user presses `q` from the main screen or `ctrl+c` from any screen.

#### Scenario: Quit from main screen
- **WHEN** the user presses `q` on the main screen
- **THEN** the TUI exits and the terminal is restored to its normal state
