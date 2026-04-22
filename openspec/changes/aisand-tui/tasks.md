## 1. Repository & Module Setup

- [x] 1.1 Verify GitHub user with `gh auth status` and create public repo `aisand` under `japacx`
- [x] 1.2 Initialize Go module `github.com/japacx/aisand` with `go mod init`
- [x] 1.3 Create full directory structure: `main.go`, `internal/ui/`, `internal/lima/`, `internal/embed/scripts/`
- [x] 1.4 Add dependencies: `charmbracelet/bubbletea`, `charmbracelet/lipgloss`, `charmbracelet/bubbles` via `go get`
- [x] 1.5 Create `.gitignore` (Go standard + binary artifacts)
- [x] 1.6 Create `README.md` with install instructions (`go install github.com/japacx/aisand@latest`) and usage

## 2. Embedded Assets

- [x] 2.1 Write updated `internal/embed/opencode-agent.yaml` — Ubuntu 25.10, home mount read-only, `runcmd` cloud-init block
- [x] 2.2 Cloud-init `runcmd`: install system deps (curl, git, build-essential, procps, file)
- [x] 2.3 Cloud-init `runcmd`: install Linuxbrew non-interactively (`NONINTERACTIVE=1`)
- [x] 2.4 Cloud-init `runcmd`: install Python 3 + pip + venv via apt
- [x] 2.5 Cloud-init `runcmd`: install Go LTS via official tarball or snap
- [x] 2.6 Cloud-init `runcmd`: install Node.js LTS via NodeSource setup script
- [x] 2.7 Cloud-init `runcmd`: configure PATH in `/etc/environment` for brew, go, node
- [x] 2.8 Write updated `internal/embed/scripts/opencode.sh` — remove brew install, keep opencode + symlinks only
- [x] 2.9 Create `internal/embed/assets.go` with `//go:embed` directives for YAML and script

## 3. Lima Client (`internal/lima/`)

- [x] 3.1 Define `VM` struct in `models.go`: Name, Status, CPUs, Memory, Disk, Mounts
- [x] 3.2 Define `Mount` struct: Location, Writable bool
- [x] 3.3 Implement `ListVMs() ([]VM, error)` — exec `limactl list --format json`, parse JSON
- [x] 3.4 Implement `StartVM(name string) *exec.Cmd` — returns command for streaming
- [x] 3.5 Implement `StopVM(name string) error` — exec `limactl stop <name>`
- [x] 3.6 Implement `DeleteVM(name string) error` — stop if running, then `limactl delete <name>`
- [x] 3.7 Implement `CreateVM(name string, cpus, ram, disk int, mounts []string, templatePath string) *exec.Cmd` — returns command for streaming
- [x] 3.8 Implement `ShellVM(name string) *exec.Cmd` — returns `limactl shell <name>` for `tea.ExecProcess`
- [x] 3.9 Implement `AddMount(name, path string) error` — `limactl edit <name> --mount <path>:w --tty=false`
- [x] 3.10 Implement `RemoveMount(name string, remainingMounts []Mount) error` — `limactl edit` with `--mount-only` or `--mount-none`
- [x] 3.11 Implement `InstallTool(name string, scriptContent []byte) *exec.Cmd` — pipes script to `limactl shell <name> bash -s`
- [x] 3.12 Implement `IsLimactlInstalled() bool` — checks PATH for `limactl`
- [x] 3.13 Implement `IsBrewInstalled() bool` — checks PATH for `brew`
- [x] 3.14 Implement `GetHostCPUs() int` and `GetHostRAMGB() int` — via `sysctl`
- [x] 3.15 Write unit tests for JSON parsing in `ListVMs` using mock output

## 4. TUI Styles (`internal/ui/styles.go`)

- [x] 4.1 Define color palette (selected item, status Running/Stopped, borders, status bar)
- [x] 4.2 Define lipgloss styles: panel border, selected row, title, status badge, log line, error text
- [x] 4.3 Define layout constants: left panel width ratio (35%), min terminal width/height

## 5. Root App Model (`internal/ui/app.go`)

- [x] 5.1 Define `AppState` enum: `StateSetup`, `StateMain`, `StateActionMenu`, `StateGlobalMenu`, `StateCreateVM`, `StateLogView`, `StateConfirm`, `StateMountsList`, `StateUnmountPicker`
- [x] 5.2 Implement `App` model with current screen, terminal size, and shared Lima client
- [x] 5.3 Implement `Init()` — check `limactl` presence, set initial state
- [x] 5.4 Implement `Update()` — delegate to current screen, handle `tea.WindowSizeMsg`, handle `ctrl+c`
- [x] 5.5 Implement `View()` — delegate to current screen

## 6. Setup Screen (`internal/ui/setup.go`)

- [x] 6.1 Implement `SetupModel` — shows missing deps, offers install option
- [x] 6.2 Display what is missing: brew (if absent) and limactl
- [x] 6.3 On user confirmation, run Homebrew install script with streaming log
- [x] 6.4 After brew, run `brew install lima` with streaming log
- [x] 6.5 On success, prompt for default resource configuration (CPUs/RAM/disk)
- [x] 6.6 Save defaults to `~/.lima/_config/default.yaml`
- [x] 6.7 On completion, transition to `StateMain`

## 7. Main Screen (`internal/ui/main.go`)

- [x] 7.1 Implement `MainModel` with left panel (VM list) and right panel (VM detail)
- [x] 7.2 Left panel: render VM list using `bubbles/list`, each item shows name + status badge
- [x] 7.3 Right panel: render selected VM detail — name, status, CPUs, RAM, disk, mounts table
- [x] 7.4 Handle empty state: no VMs — show "No VMs found. Press g → New VM to create one."
- [x] 7.5 Handle `↑`/`↓` navigation — update selected VM, refresh right panel
- [x] 7.6 Handle `enter` — transition to `StateActionMenu` with selected VM
- [x] 7.7 Handle `g` — transition to `StateGlobalMenu`
- [x] 7.8 Handle `r` — reload VM list from `limactl list`
- [x] 7.9 Handle `q` — quit
- [x] 7.10 Render status bar with keybindings: `↑↓ navigate  enter actions  g global  r refresh  q quit`

## 8. Action Menu (`internal/ui/actionmenu.go`)

- [x] 8.1 Implement `ActionMenuModel` — receives selected VM, renders numbered action list
- [x] 8.2 Actions: Shell, Start, Stop, Delete, Install tool, Mount, Unmount, Mounts
- [x] 8.3 Disable Start if VM is Running; disable Stop if VM is Stopped (show grayed out)
- [x] 8.4 Handle `enter` on Shell → `tea.ExecProcess(limaClient.ShellVM(name))`; if VM stopped, start first
- [x] 8.5 Handle `enter` on Start → transition to `StateLogView` with start command
- [x] 8.6 Handle `enter` on Stop → execute stop, refresh list, return to main
- [x] 8.7 Handle `enter` on Delete → transition to `StateConfirm` with delete callback
- [x] 8.8 Handle `enter` on Install tool → transition to `StateLogView` with install command
- [x] 8.9 Handle `enter` on Mount → transition to mount path input form
- [x] 8.10 Handle `enter` on Unmount → transition to `StateUnmountPicker`
- [x] 8.11 Handle `enter` on Mounts → transition to `StateMountsList`
- [x] 8.12 Handle `esc` → return to `StateMain`

## 9. Global Menu (`internal/ui/globalmenu.go`)

- [x] 9.1 Implement `GlobalMenuModel` — renders 4 options: New VM, Stop all VMs, Host setup, Refresh
- [x] 9.2 Handle New VM → transition to `StateCreateVM`
- [x] 9.3 Handle Stop all VMs → transition to `StateConfirm` with stop-all callback; show "No VMs running" if none
- [x] 9.4 Handle Host setup → transition to `StateSetup` (resource config flow only if Lima already installed)
- [x] 9.5 Handle Refresh → reload VM list, return to `StateMain`
- [x] 9.6 Handle `esc` → return to `StateMain`

## 10. Create VM Form (`internal/ui/createvm.go`)

- [x] 10.1 Implement multi-step `CreateVMModel` with steps: Name → CPUs → RAM → Disk → Mounts → Summary → Creating
- [x] 10.2 Step Name: text input with validation (non-empty, no spaces, not already existing in VM list)
- [x] 10.3 Step CPUs: numbered list (1 to min(hostCPUs/2, 4)), show default highlighted
- [x] 10.4 Step RAM: options 1/2/4/8 GB, show default highlighted
- [x] 10.5 Step Disk: options 20/40/60/80/100 GB, show default (60) highlighted
- [x] 10.6 Step Mounts: repeating path input — validate path exists, detect duplicates, Enter with empty path to finish
- [x] 10.7 Step Summary: show all selections, confirm [y/N]
- [x] 10.8 On confirm: write embedded YAML to temp file, build `limactl create` command, transition to `StateLogView`
- [x] 10.9 After log view completes successfully: delete temp file, refresh VM list, select new VM
- [x] 10.10 Handle `esc` at any step → go back one step (or cancel at first step)

## 11. Log View (`internal/ui/logview.go`)

- [x] 11.1 Implement `LogViewModel` — receives an `*exec.Cmd` and a title string
- [x] 11.2 Start command in goroutine, read stdout+stderr line by line via pipes
- [x] 11.3 Send each line as `LogLineMsg` to the TUI update loop
- [x] 11.4 Render lines in a `bubbles/viewport` component, auto-scroll to bottom
- [x] 11.5 On process exit code 0: show `✓ Done. Press any key to continue.`
- [x] 11.6 On process exit code != 0: show `✗ Failed (exit <code>). Press any key to continue.`
- [x] 11.7 On any key press after completion: call `onDone(exitCode)` callback to transition back

## 12. Confirm Dialog (`internal/ui/confirm.go`)

- [x] 12.1 Implement `ConfirmModel` — receives message string and `onConfirm`/`onCancel` callbacks
- [x] 12.2 Render modal overlay with message, [Y] and [N] options
- [x] 12.3 Handle `y`/`Y` → call `onConfirm()`
- [x] 12.4 Handle `n`/`N` or `esc` → call `onCancel()`

## 13. Mounts List & Unmount Picker

- [x] 13.1 Implement `MountsListModel` — renders all mounts for a VM with path and mode
- [x] 13.2 Show "No additional mounts configured" when mount list is empty
- [x] 13.3 Handle `esc` → return to `StateActionMenu`
- [x] 13.4 Implement `UnmountPickerModel` — numbered list of mounts, user selects one
- [x] 13.5 On selection: if VM running, show restart warning via `ConfirmModel`
- [x] 13.6 On confirm: stop VM (if running), call `lima.RemoveMount`, start VM (if was running)
- [x] 13.7 Handle `esc` → return to `StateActionMenu`

## 14. Mount Add Flow

- [x] 14.1 Implement mount path input screen (reuse `bubbles/textinput`)
- [x] 14.2 Validate path exists on host filesystem
- [x] 14.3 Validate path not already mounted in VM
- [x] 14.4 If VM running: show restart warning via `ConfirmModel`
- [x] 14.5 On confirm: stop VM (if running), call `lima.AddMount`, start VM (if was running)
- [x] 14.6 Refresh VM detail panel after mount is applied

## 15. Entry Point (`main.go`)

- [x] 15.1 Parse `--version` flag and print version string
- [x] 15.2 Initialize Lima client
- [x] 15.3 Create and run `App` model with `tea.NewProgram(app, tea.WithAltScreen())`
- [x] 15.4 Handle program error and exit with non-zero code on failure

## 16. Integration Testing

- [x] 16.1 Verify `go build ./...` compiles without errors
- [x] 16.2 Verify `go vet ./...` passes with no warnings
- [x] 16.3 Run `go test ./internal/lima/...` — unit tests for JSON parsing and command construction
- [ ] 16.4 Manual smoke test: run `./aisand` — verify startup check works (with and without limactl)
- [ ] 16.5 Manual smoke test: create a VM via the TUI, verify it appears in `limactl list`
- [ ] 16.6 Manual smoke test: start, stop, shell, delete a VM via the TUI
- [ ] 16.7 Manual smoke test: add and remove a mount via the TUI
- [ ] 16.8 Manual smoke test: install opencode bundle via the TUI, verify log streams in real time
- [ ] 16.9 Manual smoke test: verify cloud-init provisions brew, python3, go, node in a fresh VM
- [ ] 16.10 Verify `go install github.com/japacx/aisand@latest` installs the binary correctly after push

## 17. GitHub Repository

- [ ] 17.1 Verify GitHub auth with `gh auth status`
- [ ] 17.2 Create public repo `aisand` under `japacx` with `gh repo create japacx/aisand --public`
- [ ] 17.3 Initialize git, add all files, create initial commit
- [ ] 17.4 Push to `github.com/japacx/aisand`
- [ ] 17.5 Upload `openspec/` directory to the repo (specs committed alongside code)
- [ ] 17.6 Verify `go install github.com/japacx/aisand@latest` works from the published repo
