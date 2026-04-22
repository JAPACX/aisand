# aisand

**aisand** (AI Sandbox) is a terminal UI for managing [Lima](https://github.com/lima-vm/lima) virtual machines on macOS — purpose-built to give AI coding agents isolated Linux sandboxes where they can install dependencies, run code, and experiment freely without any risk to the host machine.

---

## Why

AI agents like [OpenCode](https://opencode.ai) need a place to act. Running them directly on your Mac means they can modify your files, install packages globally, or break your environment. aisand solves this by giving every agent its own disposable Linux VM: a full Ubuntu environment with its own filesystem, network, and tools — completely isolated from your host.

When the agent is done, you delete the VM. Your Mac is untouched.

---

## How it works

aisand wraps [limactl](https://github.com/lima-vm/lima) — the macOS VM manager built on QEMU/Virtualization.framework — with a keyboard-driven TUI. It embeds a pre-configured Lima template and installation scripts directly in the binary, so there are no external files to manage.

```
┌─────────────────────────────────────────────────────────┐
│  aisand TUI                                             │
│                                                         │
│  ┌─────────────────┐  ┌──────────────────────────────┐  │
│  │ VMs             │  │ agent-01                     │  │
│  │                 │  │ Status:  ● Running           │  │
│  │ ▶ agent-01      │  │ CPUs:    2                   │  │
│  │   agent-02      │  │ RAM:     4 GB                │  │
│  │   agent-03      │  │ Disk:    60 GB               │  │
│  │                 │  │ Mounts:  /Users/you (ro)     │  │
│  └─────────────────┘  └──────────────────────────────┘  │
│  ↑↓ navigate  enter actions  g global  r refresh  q quit│
└─────────────────────────────────────────────────────────┘
```

---

## Features

- **Isolated sandboxes** — each VM is a full Ubuntu 25.10 environment, isolated from your Mac
- **One-command install** — `go install github.com/japacx/aisand@latest`, no config files
- **Split-panel UI** — browse VMs on the left, inspect details on the right
- **Full VM lifecycle** — create, start, stop, delete, shell access — all from the keyboard
- **Real-time log streaming** — watch VM creation and tool installation output line by line
- **Tool installation** — install [OpenCode](https://opencode.ai) inside any VM with one keypress
- **Host directory mounts** — share folders from your Mac into VMs (read-only by default)
- **Host setup included** — detects missing Homebrew or Lima and offers to install them
- **Embedded assets** — Lima template and scripts are compiled into the binary

---

## Requirements

- macOS (Lima is macOS-only)
- Homebrew — installed automatically on first run if missing
- Lima (`limactl`) — installed automatically on first run if missing

---

## Install

```sh
go install github.com/japacx/aisand@latest
```

Or build from source:

```sh
git clone https://github.com/japacx/aisand
cd aisand
go build -o aisand .
./aisand
```

---

## Usage

```sh
aisand
```

On first run, aisand checks for Homebrew and Lima. If either is missing, it walks you through installation before showing the main interface.

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate VM list |
| `enter` | Open VM actions |
| `g` | Global menu (new VM, stop all, host setup) |
| `r` | Refresh VM list |
| `q` | Quit |
| `esc` | Go back |

### VM actions

| Action | Description |
|--------|-------------|
| Shell | Open an interactive shell inside the VM |
| Start | Boot the VM |
| Stop | Shut down the VM |
| Delete | Destroy the VM permanently |
| Install tool | Install tools inside the VM (e.g. OpenCode) |
| Mount | Add a host directory mount |
| Unmount | Remove a host directory mount |
| Mounts | List all active mounts |

---

## Typical workflow

```sh
# 1. Launch aisand
aisand

# 2. Press g → New VM → follow the wizard
#    (name, CPUs, RAM, disk, optional mounts)

# 3. VM starts automatically after creation

# 4. Press enter → Install tool → opencode
#    (installs OpenCode inside the VM via curl)

# 5. Press enter → Shell
#    (drops you into the VM — run opencode here)

# 6. When done, press enter → Delete
#    (VM is gone, your Mac is clean)
```

---

## Project structure

```
aisand/
├── main.go                      # Entry point
├── internal/
│   ├── ui/                      # All TUI screens (Bubbletea models)
│   │   ├── app.go               # Root model and screen router
│   │   ├── main.go              # Split-panel VM list screen
│   │   ├── actionmenu.go        # Per-VM action menu
│   │   ├── globalmenu.go        # Global operations menu
│   │   ├── createvm.go          # Multi-step VM creation wizard
│   │   ├── logview.go           # Real-time log streaming panel
│   │   ├── confirm.go           # Confirmation dialog
│   │   ├── setup.go             # Host setup / onboarding screen
│   │   ├── toolpicker.go        # Tool installation picker
│   │   ├── mountslist.go        # Mounts viewer and unmount picker
│   │   ├── mountinput.go        # Mount path input
│   │   └── styles.go            # Shared lipgloss styles
│   ├── lima/
│   │   ├── client.go            # limactl CLI wrapper
│   │   └── models.go            # VM and Mount structs
│   └── embed/
│       ├── assets.go            # //go:embed directives
│       ├── opencode-agent.yaml  # Lima VM template
│       └── scripts/
│           └── opencode.sh      # OpenCode installation script
└── openspec/                    # Design specs and task history
```

---

## Tech stack

| Component | Library |
|-----------|---------|
| TUI framework | [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) |
| Styling | [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) |
| UI components | [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) |
| VM backend | [lima-vm/lima](https://github.com/lima-vm/lima) via `limactl` CLI |

---

## License

MIT
