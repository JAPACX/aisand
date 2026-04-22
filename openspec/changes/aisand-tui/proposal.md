## Why

Managing Lima VMs today requires memorizing and running Makefile commands manually — there is no visual feedback, no way to see all VMs and their state at a glance, and every operation requires switching to a terminal and typing the right incantation. The goal is to replace this entirely with a single installable Go binary (`aisand`) that provides a full TUI for Lima VM management, embeds all required templates and scripts, and is distributable via `go install github.com/japacx/aisand@latest`.

## What Changes

- **NEW**: `aisand` Go binary — a TUI application built with Bubbletea that replaces the Makefile entirely
- **NEW**: Split-panel main screen — left panel lists all VMs with status, right panel shows selected VM detail (CPUs, RAM, disk, mounts)
- **NEW**: Contextual action menu per VM (Enter) — Shell, Start, Stop, Delete, Install tool, Mount, Unmount, Mounts list
- **NEW**: Global menu (key `g`) — New VM, Stop all VMs, Host setup, Refresh
- **NEW**: Onboarding/setup screen — blocks all functionality if `limactl` is not installed; guides user through Homebrew + Lima installation
- **NEW**: Real-time log panel — streaming stdout/stderr for long operations (VM creation, opencode install, ~5-10 min)
- **NEW**: `opencode-agent.yaml` updated with cloud-init provisioning — Linuxbrew, Python 3, Go LTS, Node.js LTS pre-installed at VM boot
- **MODIFIED**: `opencode.sh` — brew installation removed (now handled by cloud-init); script only installs opencode + symlinks
- **REMOVED**: `vm:addon` functionality — not included in the TUI
- **NEW**: GitHub repo `github.com/japacx/aisand` — public, with all source code and openspec specs

## Capabilities

### New Capabilities

- `tui-shell`: Main TUI application shell — startup check, layout engine, navigation model, keybindings
- `vm-management`: All VM lifecycle operations — list, create, start, stop, delete, shell access
- `vm-mounts`: Volume management — add host directory mounts, remove mounts, list mounts per VM
- `tool-install`: Tool bundle installation inside VMs — opencode bundle with real-time log streaming
- `host-setup`: Host dependency management — detect, install and configure Homebrew + Lima on macOS
- `vm-template`: Embedded Lima YAML template with cloud-init provisioning (brew + Go + Node + Python)

### Modified Capabilities

_(none — this is a greenfield binary; the Makefile is being replaced, not modified)_

## Impact

- **New repo**: `github.com/japacx/aisand` (public) — all source code, go.mod, embedded assets
- **Go module**: `github.com/japacx/aisand` — installable via `go install github.com/japacx/aisand@latest`
- **Dependencies**: `charmbracelet/bubbletea`, `charmbracelet/lipgloss`, `charmbracelet/bubbles`
- **Embedded assets**: `opencode-agent.yaml` (Lima template), `scripts/opencode.sh` (tool bundle) — compiled into the binary via `//go:embed`
- **External dependency**: `limactl` must be present on the host (enforced at startup)
- **macOS only**: TUI targets macOS (Lima is macOS-only); no Linux/Windows support needed
- **Existing Makefile**: Remains in the repo as reference but is superseded by the TUI
