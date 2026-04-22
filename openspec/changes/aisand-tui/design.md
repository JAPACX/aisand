## Context

The current workflow for managing Lima VMs is entirely Makefile-based. Users must know the exact `make` targets, pass arguments correctly, and interpret raw terminal output. There is no persistent view of VM state, no visual feedback during long operations, and no discoverability of available actions.

The replacement is `aisand` — a self-contained Go binary that embeds all required assets (Lima YAML template, shell scripts) and provides a full TUI. It targets macOS developers who use Lima to run isolated Linux VMs for AI agent workloads (specifically OpenCode agents).

**Current state:**
- `Makefile` with targets: `host:install`, `vm:new`, `vm:install`, `vm:start`, `vm:stop`, `vm:delete`, `vm:shell`, `vm:list`, `vm:mount`, `vm:unmount`, `vm:mounts`
- `opencode-agent.yaml` — Lima template (Ubuntu 25.10, home dir mounted read-only)
- `scripts/opencode.sh` — installs Linuxbrew + opencode + symlinks inside VM
- `scripts/addon.sh` — personal customization script (being dropped)

**Constraints:**
- Must wrap `limactl` CLI — Lima has no stable Go API, only a CLI
- Shell access requires suspending the TUI and handing control to the terminal
- Mount changes require VM restart (Lima limitation)
- `go install` requires the binary to be buildable without CGO dependencies
- All assets must be embedded in the binary (no external file dependencies at runtime)

## Goals / Non-Goals

**Goals:**
- Replace all Makefile functionality (except `vm:addon`) with a single TUI binary
- Provide split-panel layout: VM list (left) + VM detail (right)
- Contextual action menu per VM (Enter key)
- Global menu for cross-VM operations (key `g`)
- Onboarding screen that blocks usage until `limactl` is installed
- Real-time log streaming for long operations (VM create, tool install)
- Embed `opencode-agent.yaml` and `scripts/opencode.sh` in the binary
- Update `opencode-agent.yaml` to provision brew + Go + Node.js + Python via cloud-init
- Publish to `github.com/japacx/aisand` as a public Go module
- Installable via `go install github.com/japacx/aisand@latest`

**Non-Goals:**
- Linux or Windows support (Lima is macOS-only)
- `vm:addon` functionality
- A web UI or REST API
- Managing Lima instances on remote machines
- Supporting Lima templates other than the embedded one
- Automatic updates of the binary

## Decisions

### Decision: Bubbletea as TUI framework

**Choice**: `charmbracelet/bubbletea` + `charmbracelet/lipgloss` + `charmbracelet/bubbles`

**Rationale**: Bubbletea is the de-facto standard for Go TUIs. It uses the Elm architecture (Model/Update/View) which makes complex stateful UIs manageable. Lipgloss handles layout and styling. Bubbles provides ready-made components (list, spinner, viewport, textinput) that map directly to the required UI elements.

**Alternatives considered**:
- `tview` — more widget-based, less composable, harder to style
- `termui` — lower-level, more boilerplate for the same result
- Raw `tcell` — too low-level for this scope

---

### Decision: Wrap `limactl` CLI via `exec.Command`

**Choice**: All Lima operations go through `os/exec` calls to `limactl`.

**Rationale**: Lima does not expose a stable Go API. The CLI is the only supported interface. Parsing `limactl list --format json` gives structured VM data. For streaming output (create, install), `cmd.StdoutPipe()` + `cmd.StderrPipe()` feed directly into the TUI log panel via a channel.

**Alternatives considered**:
- Lima Go internals — unstable, not exported, would break on Lima updates
- gRPC/socket — Lima does not expose one

---

### Decision: `//go:embed` for assets

**Choice**: `opencode-agent.yaml` and `scripts/opencode.sh` are embedded using `//go:embed` directives in an `internal/embed` package.

**Rationale**: The binary must be self-contained. Users install it with `go install` and should not need to clone the repo or manage external files. Embedding ensures the correct template and script version ships with each binary release.

**Implementation**: At VM creation time, the embedded YAML is written to a temp file, passed to `limactl create --template <tmpfile>`, then deleted. For tool install, the embedded script is piped to `limactl shell <name> bash -s`.

---

### Decision: Shell access via TUI suspension

**Choice**: When the user selects "Shell" for a VM, the TUI calls `tea.ExecProcess()` to suspend Bubbletea, hand control to `limactl shell <name>`, and resume the TUI when the shell exits.

**Rationale**: Interactive shells require raw terminal control that conflicts with Bubbletea's event loop. `tea.ExecProcess` is the official Bubbletea mechanism for this pattern — it cleanly suspends the TUI, runs the subprocess with full terminal control, and resumes.

---

### Decision: Cloud-init for VM base provisioning

**Choice**: `opencode-agent.yaml` gains a `runcmd` section that installs Linuxbrew, Go, Node.js LTS, and Python 3 at first boot.

**Rationale**: Previously, brew was installed by `opencode.sh` at tool-install time. Moving it to cloud-init means the VM is ready for any tool installation immediately after `limactl start` completes, without a separate provisioning step. This also makes the `opencode.sh` script simpler and faster.

**Cloud-init sequence**:
1. `apt-get install -y curl git build-essential procps file` — system deps
2. `NONINTERACTIVE=1 bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"` — Linuxbrew
3. `apt-get install -y python3 python3-pip python3-venv` — Python 3
4. `snap install go --classic` OR official Go tarball — Go LTS
5. NodeSource setup script + `apt-get install -y nodejs` — Node.js LTS
6. PATH configuration in `/etc/environment`

---

### Decision: Application state machine

**Choice**: The TUI is modeled as a state machine with explicit screen states.

**States**:
```
SetupRequired   → user must install Lima before proceeding
MainScreen      → split panel: VM list + detail
ActionMenu      → contextual menu for selected VM
GlobalMenu      → cross-VM operations
CreateVMForm    → multi-step form: name → CPUs → RAM → disk → mounts
LogView         → streaming output for long operations
ConfirmDialog   → destructive action confirmation (delete, stop-all)
MountsList      → list of mounts for a VM
UnmountPicker   → select mount to remove
```

**Transitions**: Each screen is a Bubbletea `Model`. The root `App` model holds the current screen and delegates `Update`/`View` to it. Screen transitions happen by replacing the current screen in the App model.

---

### Decision: Real-time log streaming architecture

**Choice**: Long operations run in a goroutine. Output lines are sent to the TUI via a `tea.Cmd` that returns `tea.Msg` values. The `LogView` screen appends lines to a `viewport` component.

**Implementation**:
```
goroutine: exec limactl → read stdout/stderr line by line → send LogLineMsg
TUI Update: receive LogLineMsg → append to viewport → scroll to bottom
goroutine: on exit → send OperationDoneMsg (with exit code)
TUI Update: receive OperationDoneMsg → show result + "Press any key to continue"
```

---

### Decision: Project structure

```
aisand/
├── main.go                         # Entry point, version flag
├── go.mod                          # module github.com/japacx/aisand
├── go.sum
├── internal/
│   ├── ui/
│   │   ├── app.go                  # Root model, screen router
│   │   ├── setup.go                # SetupRequired screen
│   │   ├── main.go                 # MainScreen (split panel)
│   │   ├── actionmenu.go           # VM contextual menu
│   │   ├── globalmenu.go           # Global menu
│   │   ├── createvm.go             # Multi-step VM creation form
│   │   ├── logview.go              # Streaming log panel
│   │   ├── confirm.go              # Confirmation dialog
│   │   ├── mountslist.go           # Mounts viewer
│   │   ├── unmountpicker.go        # Mount removal picker
│   │   └── styles.go               # Lipgloss styles (shared)
│   ├── lima/
│   │   ├── client.go               # limactl exec wrapper
│   │   └── models.go               # VM, Mount structs
│   └── embed/
│       ├── assets.go               # //go:embed directives
│       ├── opencode-agent.yaml     # Lima template with cloud-init
│       └── scripts/
│           └── opencode.sh         # Tool bundle (opencode only)
└── openspec/                       # Specs (committed to repo)
```

## Risks / Trade-offs

**[Risk] `limactl` CLI output format changes** → Mitigation: Pin to `limactl list --format json` which is more stable than human-readable output. Add integration test that validates JSON parsing against a real `limactl list` call.

**[Risk] Cloud-init provisioning takes 10-15 min on first boot** → Mitigation: The TUI shows a log panel with real-time cloud-init output so the user can see progress. Document expected duration in the UI.

**[Risk] `tea.ExecProcess` behavior differences across terminal emulators** → Mitigation: Test in iTerm2 and Terminal.app. This is a known-good Bubbletea pattern used in production tools (e.g., Charm's own tools).

**[Risk] Mount changes require VM restart** → Mitigation: The TUI explicitly warns the user before applying mount changes and handles the stop/edit/start cycle automatically, matching the Makefile behavior.

**[Risk] `go install` requires all dependencies to be available** → Mitigation: No CGO, no system libraries. Pure Go + embedded assets. The binary will cross-compile cleanly.

**[Risk] Embedded `opencode-agent.yaml` becomes stale** → Mitigation: The YAML is versioned in the repo. Users can `go install` the latest version to get updated templates. Document this in README.
