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

## Project structure

```
internal/ui/       — TUI screens (Bubbletea models)
internal/lima/     — limactl CLI wrapper and VM types
internal/embed/    — embedded Lima template and scripts
openspec/          — design specs, decisions, and task history
```

Each TUI screen is a self-contained Bubbletea `Model` (Init / Update / View). Screen transitions happen via `ChangeScreenMsg` sent to the root `App` model in `app.go`.

---

## Development workflow

### Making changes

1. Fork the repo and create a branch from `main`
2. Make your changes — keep them focused and minimal
3. Ensure `go build ./...` and `go vet ./...` pass with no output
4. Run `go test ./...` — all tests must pass
5. Test manually with `go run .`

### Commit style

Use conventional commits:

```
feat: add support for VM snapshots
fix: prevent escape key from exiting log view during operation
refactor: extract VM reload logic into shared helper
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
- **Additional tool bundles** — add new scripts to `internal/embed/scripts/` and register them in `toolpicker.go` (e.g. Claude Code, Cursor, other agents)
- **VM templates** — alternative Lima YAML templates (different Ubuntu versions, different resource profiles)
- **Test coverage** — unit tests for UI models using [bubbletea testing utilities](https://github.com/charmbracelet/bubbletea/tree/main/teatest)
- **Error handling** — better error messages when limactl operations fail

### Medium value
- **VM renaming** — Lima supports it via `limactl edit`
- **Resource editing** — change CPUs/RAM/disk on existing VMs
- **VM status polling** — auto-refresh the VM list while operations are in progress

### Not in scope
- Linux or Windows support (Lima is macOS-only)
- A web UI or REST API
- Managing Lima instances on remote machines
- Supporting non-embedded Lima templates at runtime
- Automatic binary updates

---

## Adding a new tool

Tools are registered in `internal/ui/toolpicker.go`. To add a new tool:

1. Add the installation script to `internal/embed/scripts/<toolname>.sh`
   - The script must be idempotent (safe to run multiple times)
   - It must work in a non-interactive `bash -s` session (no TTY)
   - Check for the tool before installing, print `[✓] already installed` if found

2. Add an embed directive in `internal/embed/assets.go`:
   ```go
   //go:embed scripts/toolname.sh
   var ToolNameScript []byte
   ```

3. Register it in `availableTools` in `toolpicker.go`:
   ```go
   {
       name:        "toolname",
       description: "Short description of what it does",
       script:      func() []byte { return embed.ToolNameScript },
   },
   ```

4. Add a check in `checkToolsCmd` so the picker shows installed/not-installed status.

---

## Code style

- Standard Go formatting — `gofmt` before committing
- No external dependencies beyond the existing ones without discussion
- Keep files focused — one screen per file in `internal/ui/`
- Prefer explicit over clever — this codebase is meant to be readable

---

## Reporting issues

Open a GitHub issue with:
- What you did
- What you expected
- What happened instead
- Your macOS version and Lima version (`limactl --version`)

---

## Questions

Open a GitHub discussion or issue. We are happy to discuss design decisions before you invest time in a large change.
