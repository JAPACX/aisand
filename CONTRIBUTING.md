# Contributing to aisand

Thank you for your interest in contributing. aisand is a focused tool — a TUI for managing Lima VMs as AI agent sandboxes on macOS. Contributions that improve reliability, usability, or expand the sandbox use case are welcome.

---

## Philosophy

aisand exists to solve one problem well: give AI agents isolated Linux environments on macOS with zero friction. Every contribution should serve that goal. Features that add complexity without clear benefit to the sandbox workflow will not be merged.

---

## Getting started

### Prerequisites

- macOS
- Go 1.21+
- Homebrew
- Lima (`brew install lima`)

### Setup

```sh
git clone https://github.com/japacx/aisand
cd aisand
go mod download
go build ./...
go test ./...
```

### Run locally

```sh
go run .
```

---

## Architecture

### Screen state machine

The TUI is modeled as a state machine. The root `App` model in `app.go` holds the current screen as a `tea.Model` interface and delegates `Update`/`View` to it. Screen transitions are triggered by sending a `ChangeScreenMsg` from any screen.

```
StateSetup → StateMain → StateActionMenu → StateLogView
                      ↘ StateGlobalMenu → StateCreateVM → StateLogView
                                        ↘ StateConfirm
                                        ↘ StateToolPicker → StateLogView
                                        ↘ StateMountsList
                                        ↘ StateUnmountPicker → StateConfirm → StateLogView
```

### Operation safety model

- All in-progress operations (log view) block all keyboard input
- Exception: VM creation allows `esc` → cancel confirm → kill process + delete VM
- If user says "No" to cancel, the original `LogViewModel` is restored via `resumeLogViewMsg` handled in `App.Update`
- VM actions that require Running status are disabled and skipped by arrow navigation

### Tool installation

Tools are installed natively inside the VM via `limactl shell <name> bash -s`. The script is piped via stdin. macOS binaries cannot be shared with the VM — the kernel ABI (Mach-O vs ELF) is incompatible regardless of architecture. Config files and credentials work via symlinks because they are plain text, not executables.

### Credential sharing

The VM mounts the host home directory read-only via virtiofs. Install scripts create symlinks inside the VM pointing to the host config paths. The VM reads credentials from the Mac without copying or modifying them.

---

## Project structure

```
internal/ui/       — TUI screens (Bubbletea models, one file per screen)
internal/lima/     — limactl CLI wrapper and VM/Mount types
internal/embed/    — embedded Lima YAML template and install scripts
openspec/          — design specs, decisions, and full task history
```

---

## Development workflow

### Making changes

1. Fork the repo and create a branch from `main`
2. Make your changes — keep them focused and minimal
3. Ensure `go build ./...` and `go vet ./...` pass with no output
4. Run `go test ./...` — all tests must pass
5. Test manually with `go run .`

### Versioning

The version string is hardcoded in `main.go`. When releasing:

1. Update `const version = "x.y.z"` in `main.go`
2. Commit with `fix: bump version to x.y.z`
3. Create a git tag `vx.y.z` pointing to that commit

Users install via `GOPROXY=direct go install github.com/japacx/aisand@latest` to bypass proxy cache.

### Commit style

Use conventional commits:

```
feat: add VM snapshot support
fix: disable Shell action when VM is Stopped
refactor: extract VM reload into backToAction helper
docs: update contributing guide
```

### Pull requests

- One concern per PR — do not bundle unrelated changes
- Describe what the change does and why
- If it changes behavior, describe the before/after
- PRs that break `go build`, `go vet`, or `go test` will not be reviewed

---

## Areas open for contribution

### High value

- **Additional tool bundles** — add scripts to `internal/embed/scripts/` and register in `toolpicker.go` (e.g. Claude Code, other AI agents)
- **VM templates** — alternative Lima YAML templates (different Ubuntu versions, minimal profiles)
- **Test coverage** — unit tests for UI models using [teatest](https://github.com/charmbracelet/bubbletea/tree/main/teatest)
- **Error handling** — better error messages when limactl operations fail

### Medium value

- **VM resource editing** — change CPUs/RAM/disk on existing VMs via `limactl edit`
- **VM status auto-refresh** — poll `limactl list` periodically while a VM is starting
- **Multiple tool scripts** — install multiple tools in one pass

### Not in scope

- Linux or Windows support (Lima is macOS-only)
- A web UI or REST API
- Managing Lima instances on remote machines
- Sharing macOS binaries with VMs (not technically feasible — ABI incompatibility)
- Automatic binary self-updates

---

## Adding a new tool

Tools are registered in `internal/ui/toolpicker.go`. To add one:

**1. Write the install script** at `internal/embed/scripts/<toolname>.sh`

Requirements:
- Must be idempotent — safe to run multiple times
- Must work in a non-interactive `bash -s` session (no TTY, no prompts)
- Must export the brew PATH at the top if it uses brew: `export PATH="/home/linuxbrew/.linuxbrew/bin:..."`
- Must check if the tool is already installed before installing
- Must print clear status lines (`[✓] already installed`, `[✓] installed`, etc.)
- Must install the Linux-native binary — do not attempt to symlink Mac binaries

**2. Add an embed directive** in `internal/embed/assets.go`:

```go
//go:embed scripts/toolname.sh
var ToolNameScript []byte
```

**3. Register it** in `availableTools` in `toolpicker.go`:

```go
{
    name:        "toolname",
    description: "Short description of what it does",
    script:      func() []byte { return embed.ToolNameScript },
},
```

**4. Add a status check** in `checkToolsCmd` in `toolpicker.go`:

```sh
if command -v toolname &>/dev/null; then echo "toolname:installed"; else echo "toolname:missing"; fi
```

The status parser in `checkToolsCmd` is generic — it iterates `availableTools` and matches `name:installed` / `name:missing` lines, so no additional parsing code is needed.

---

## Code style

- Standard Go formatting — run `gofmt` before committing
- No new external dependencies without discussion
- One screen per file in `internal/ui/`
- Prefer explicit over clever — this codebase is meant to be readable by anyone familiar with Bubbletea

---

## Reporting issues

Open a GitHub issue with:
- What you did
- What you expected
- What happened instead
- macOS version, Lima version (`limactl --version`), aisand version (`aisand --version`)

---

## Questions

Open a GitHub discussion or issue before investing time in a large change — design decisions matter here.
