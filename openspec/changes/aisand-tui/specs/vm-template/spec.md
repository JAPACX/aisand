## ADDED Requirements

### Requirement: Embedded Lima template
The system SHALL embed the `opencode-agent.yaml` Lima template inside the binary using `//go:embed`. The template SHALL be written to a temporary file at VM creation time, passed to `limactl create`, and deleted after the command completes.

#### Scenario: VM creation uses embedded template
- **WHEN** the user creates a new VM
- **THEN** the system writes the embedded YAML to a temp file
- **THEN** passes it to `limactl create --name <name> <tmpfile>`
- **THEN** deletes the temp file after the command exits

---

### Requirement: Cloud-init base provisioning
The embedded Lima template SHALL include a `runcmd` cloud-init section that installs the following on first VM boot, without any manual intervention: Linuxbrew (non-interactive), Python 3 with pip and venv, Go (latest LTS), Node.js LTS (via NodeSource). All tools SHALL be available in PATH for all subsequent shell sessions.

#### Scenario: VM first boot provisioning
- **WHEN** a new VM starts for the first time
- **THEN** cloud-init runs the provisioning commands automatically
- **THEN** after provisioning completes, `brew`, `python3`, `go`, and `node` are available in the VM shell

#### Scenario: Provisioning is idempotent
- **WHEN** the VM is stopped and restarted
- **THEN** cloud-init does not re-run the provisioning commands
- **THEN** all tools remain available

---

### Requirement: Embedded opencode script
The system SHALL embed `scripts/opencode.sh` inside the binary using `//go:embed`. The script SHALL assume Linuxbrew is already installed (via cloud-init) and SHALL only install: opencode via `brew install anomalyco/tap/opencode`, symlink `~/.config/opencode` → host virtiofs mount, symlink `~/.local/share/opencode/auth.json` → host virtiofs mount.

#### Scenario: opencode.sh runs on provisioned VM
- **WHEN** the embedded opencode.sh is executed inside a VM that has been provisioned via cloud-init
- **THEN** the script finds `brew` available and installs opencode
- **THEN** creates the config and auth symlinks pointing to the host home via virtiofs

#### Scenario: opencode.sh is idempotent
- **WHEN** the script is run a second time on the same VM
- **THEN** each step reports `[✓] already installed` or `[✓] symlink ok`
- **THEN** no reinstallation occurs

---

### Requirement: Host home mount
The embedded Lima template SHALL mount the macOS host home directory inside the VM as read-only via virtiofs. The mount point SHALL be the host user's home path (e.g., `/Users/<username>`).

#### Scenario: Host home accessible in VM
- **WHEN** a VM is running
- **THEN** the host home directory is accessible inside the VM at the same path as on the host
- **THEN** the mount is read-only — the VM cannot modify host files
