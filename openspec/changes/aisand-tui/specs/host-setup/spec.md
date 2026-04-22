## ADDED Requirements

### Requirement: Detect host dependencies
The system SHALL check for the presence of `limactl` in PATH at startup. The system SHALL also check for `brew` when the user initiates host setup.

#### Scenario: limactl present
- **WHEN** `limactl` is found in PATH at startup
- **THEN** the system proceeds to the main screen

#### Scenario: limactl absent
- **WHEN** `limactl` is not found in PATH at startup
- **THEN** the system displays the Host Setup Required screen
- **THEN** the system blocks access to all VM management functionality

---

### Requirement: Onboarding screen
The system SHALL display a blocking onboarding screen when `limactl` is not installed. The screen SHALL explain what is missing and offer to install it. The screen SHALL show real-time output during installation.

#### Scenario: User initiates install from onboarding
- **WHEN** the user selects "Install Lima" on the onboarding screen
- **THEN** the system checks if `brew` is installed; if not, installs Homebrew first
- **THEN** the system runs `brew install lima`
- **THEN** all output is streamed in real time
- **THEN** on success the system transitions to the main screen

#### Scenario: Homebrew not installed
- **WHEN** `brew` is not found and the user confirms installation
- **THEN** the system runs the official Homebrew install script
- **THEN** proceeds to install Lima after Homebrew is ready

---

### Requirement: Configure default VM resources
The system SHALL allow the user to configure default VM resource allocations (CPUs, RAM, disk) that are saved to `~/.lima/_config/default.yaml`. This is accessible from both the onboarding screen (after Lima install) and the Host setup option in the global menu.

#### Scenario: Set defaults after Lima install
- **WHEN** Lima is successfully installed via the onboarding screen
- **THEN** the system prompts the user to configure default resources
- **THEN** saves the selection to `~/.lima/_config/default.yaml`

#### Scenario: Update defaults from global menu
- **WHEN** the user selects Host setup from the global menu and Lima is already installed
- **THEN** the system shows current defaults (if any)
- **THEN** allows the user to update CPUs, RAM, and disk defaults
- **THEN** saves the updated values to `~/.lima/_config/default.yaml`

#### Scenario: Default resource ranges
- **WHEN** the user is selecting default CPUs
- **THEN** the valid range is 1 to half of host CPUs (max 4)
- **WHEN** the user is selecting default RAM
- **THEN** the valid options are 1, 2, 4, or 8 GB
- **WHEN** the user is selecting default disk
- **THEN** the valid options are 20, 40, 60, 80, or 100 GB
