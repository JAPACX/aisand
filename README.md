# aisand

**aisand** (AI Sandbox) is a terminal UI for managing [Lima](https://github.com/lima-vm/lima) virtual machines on macOS — purpose-built to give AI coding agents isolated Linux sandboxes where they can install dependencies, run code, and experiment freely without any risk to the host machine.

---

## Why

AI agents like [OpenCode](https://opencode.ai) need a place to act. Running them directly on your Mac means they can modify your files, install packages globally, or break your environment. aisand solves this by giving every agent its own disposable Linux VM: a full Ubuntu 25.10 environment with its own filesystem, network, and tools — completely isolated from your host.

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

- **Isolated sandboxes** — each VM is a full Ubuntu 25.10 environment, completely isolated from your Mac
- **One-command install** — `go install github.com/japacx/aisand@latest`, no config files needed
- **Host setup on first run** — detects missing Homebrew or Lima and installs them before anything else
- **Split-panel UI** — browse VMs on the left, inspect details on the right
- **Full VM lifecycle** — create, start, stop, delete, and shell into VMs from the keyboard
- **Operation safety** — all in-progress operations are non-interruptible; actions unavailable for stopped VMs are grayed out
- **Real-time log streaming** — watch VM creation, start, stop, and tool installation output line by line with a live spinner
- **Tool installation** — install tools (Homebrew, OpenCode) inside VMs from a picker that shows what is already installed
- **Zero credential setup** — API keys and config symlinked automatically from your Mac into every VM
- **Host directory mounts** — share any Mac folder into VMs as read-write; aisand handles the stop/edit/start cycle automatically
- **Embedded assets** — Lima template and install scripts compiled into the binary, no external files

---

## How credentials are shared

When you install OpenCode inside a VM via aisand, the install script creates two symlinks inside the VM pointing to your Mac's config via the virtiofs host mount:

```
~/.config/opencode               → /Users/<you>/.config/opencode
~/.local/share/opencode/auth.json → /Users/<you>/.local/share/opencode/auth.json
```

This means:

- **You never copy API keys into the VM** — the VM reads them directly from your Mac
- **Every VM shares the same credentials** — configure once on your Mac, all VMs pick it up automatically
- **The VM cannot modify your config** — the host home mount is read-only

Open a shell inside any VM and run `opencode` — it is already authenticated, no extra setup needed.

> **Note on sharing binaries**: symlinks or copies of macOS binaries (e.g. `/usr/local/bin/opencode`) will **not** work inside Linux VMs. macOS binaries use the Mach-O format; Linux requires ELF. They are incompatible at the kernel level regardless of architecture. Tools must be installed natively inside the VM — that is what the Install tool flow does.

---

## Requirements

- macOS (Lima is macOS-only)
- Go 1.21+ (to install via `go install`)
- Homebrew — detected and installed automatically on first run if missing
- Lima (`limactl`) — detected and installed automatically on first run if missing

---

## Install

```sh
go install github.com/japacx/aisand@latest
```

After installing, make sure `~/go/bin` is in your PATH. If `aisand` gives you `command not found`, add this to your `~/.zshrc`:

```sh
export PATH="$HOME/go/bin:$PATH"
```

Then reload:

```sh
source ~/.zshrc
```

Or build from source:

```sh
git clone https://github.com/japacx/aisand
cd aisand
go build -o aisand .
./aisand
```

### Updating

To get the latest version:

```sh
GOPROXY=direct go install github.com/japacx/aisand@latest
```

The `GOPROXY=direct` flag bypasses the Go module proxy cache and fetches directly from GitHub.

---

## First run

On first launch, aisand checks for Homebrew and Lima. If either is missing, it shows a **Host Setup** screen and walks you through installation before showing the main interface. You cannot create or manage VMs until both dependencies are present.

If you return to Host Setup after everything is already installed, the install option is disabled and the screen shows all dependencies as satisfied.

---

## Usage

```sh
aisand
```

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate list |
| `enter` | Confirm / open actions |
| `g` | Global menu (new VM, stop all, host setup) |
| `r` | Refresh VM list |
| `q` | Quit |
| `esc` | Go back |

### VM actions

Actions available depend on VM status. Grayed-out items cannot be selected.

| Action | Requires | Description |
|--------|----------|-------------|
| Shell | Running | Open an interactive shell inside the VM |
| Start | Stopped | Boot the VM (streams output in real time) |
| Stop | Running | Shut down the VM (streams output in real time) |
| Delete | Any | Destroy the VM permanently (requires confirmation) |
| Install tool | Running | Open the tool picker to install tools inside the VM |
| Mount | Running | Add a host directory as a read-write mount |
| Unmount | Running | Remove a mount (VM is restarted automatically) |
| Mounts | Running | List all active mounts |

> All in-progress operations block navigation until complete. During VM creation, pressing `esc` asks if you want to cancel — if confirmed, the VM is deleted. For all other operations, you must wait for completion.

---

## Tool installation

The **Install tool** picker connects to the VM, checks which tools are already installed, and shows their status before you act:

```
▶ brew        [ not installed ]  Homebrew package manager (required for many tools)
  opencode    [✓ installed    ]  AI coding agent (curl install + config symlinks)
```

- Already installed tools are grayed out and cannot be selected
- Tools are installed natively inside the VM (not symlinked from the Mac — see note above)
- Scripts are embedded in the aisand binary — no internet access needed to get the scripts

### Available tools

| Tool | How installed | What it does |
|------|--------------|--------------|
| `brew` | Official Homebrew install script | Package manager for Linux inside the VM |
| `opencode` | `curl -fsSL https://opencode.ai/install \| bash` | AI coding agent with auto config symlinks |

---

## Typical workflow

```
1. Run: aisand
   → Host Setup runs automatically if brew or limactl are missing

2. Press g → New VM
   → Name, CPUs, RAM, disk, optional host folder mounts
   → VM is created and started; watch the log in real time

3. Press enter on the VM → Install tool → opencode
   → Script runs inside the VM; credentials symlinked from your Mac automatically

4. Press enter → Shell
   → You are now inside the VM; run opencode, experiment freely

5. When done → Delete
   → VM destroyed, your Mac is clean
```

---

## Project structure

```
aisand/
├── main.go                       # Entry point, version flag
├── internal/
│   ├── ui/                       # All TUI screens (Bubbletea models)
│   │   ├── app.go                # Root model, screen router, state machine
│   │   ├── setup.go              # Host setup / onboarding (brew + lima)
│   │   ├── main.go               # Split-panel VM list screen
│   │   ├── actionmenu.go         # Per-VM contextual action menu
│   │   ├── globalmenu.go         # Global operations menu
│   │   ├── createvm.go           # Multi-step VM creation wizard
│   │   ├── logview.go            # Real-time log streaming with spinner
│   │   ├── confirm.go            # Confirmation dialog (y/n only)
│   │   ├── toolpicker.go         # Tool picker with installed/missing status
│   │   ├── mountslist.go         # Mounts viewer
│   │   ├── mountinput.go         # Add mount path input
│   │   ├── unmountpicker.go      # Remove mount picker
│   │   └── styles.go             # Shared lipgloss color palette and styles
│   ├── lima/
│   │   ├── client.go             # limactl CLI wrapper (all VM operations)
│   │   └── models.go             # VM and Mount structs
│   └── embed/
│       ├── assets.go             # //go:embed directives
│       ├── opencode-agent.yaml   # Embedded Lima VM template (Ubuntu 25.10)
│       └── scripts/
│           ├── brew.sh           # Homebrew install script (idempotent)
│           └── opencode.sh       # OpenCode install + config symlinks (idempotent)
└── openspec/                     # Design specs, decisions, and task history
```

---

## Technical notes

### Why tools must be installed inside the VM

macOS binaries (Mach-O format, darwin/arm64) are not executable on Linux (ELF format, linux/arm64). Symlinks or copies of Mac binaries into the VM will fail with `exec format error` at the kernel level. Tools must be installed natively inside the VM for the correct OS and ABI.

### Why credentials work via symlinks but binaries don't

Config files and JSON auth tokens are plain text — format-agnostic. The VM reads them as files via the virtiofs mount with no OS-level interpretation. Binaries are executed by the kernel, which enforces strict format compatibility.

### limactl is the only interface

Lima does not expose a stable Go API. All VM operations go through `exec.Command("limactl", ...)`. JSON output from `limactl list --format json` is parsed as NDJSON (one object per line).

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
