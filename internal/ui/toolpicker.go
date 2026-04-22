package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/japacx/aisand/internal/embed"
	"github.com/japacx/aisand/internal/lima"
)

// tool describes an installable tool bundle.
type tool struct {
	name           string
	description    string
	script         func() []byte // returns the script to pipe to bash -s
	askSymlinks    bool          // if true, ask user about host config symlinks before installing
}

// availableTools is the registry of installable tools.
var availableTools = []tool{
	{
		name:        "brew",
		description: "Homebrew package manager (required for many tools)",
		script:      func() []byte { return embed.BrewScript },
	},
	{
		name:        "opencode",
		description: "AI coding agent (curl install + optional config symlinks)",
		script:      func() []byte { return embed.OpenCodeScript },
		askSymlinks: true,
	},
}

// toolStatusMsg carries the install status of each tool, keyed by tool name.
type toolStatusMsg map[string]bool // true = installed

// ToolPickerModel lets the user pick a tool to install and shows its status.
type ToolPickerModel struct {
	client   *lima.Client
	vm       lima.VM
	tools    []tool
	status   map[string]bool
	loading  bool
	selected int
	width    int
	height   int
}

// NewToolPickerModel creates a new ToolPickerModel.
func NewToolPickerModel(client *lima.Client, vm lima.VM, width, height int) *ToolPickerModel {
	return &ToolPickerModel{
		client:  client,
		vm:      vm,
		tools:   availableTools,
		status:  map[string]bool{},
		loading: true,
		width:   width,
		height:  height,
	}
}

// checkToolsCmd runs a quick check inside the VM to see which tools are installed.
func (m *ToolPickerModel) checkToolsCmd() tea.Cmd {
	client := m.client
	vmName := m.vm.Name
	return func() tea.Msg {
		status := map[string]bool{}
		checkScript := `
export PATH="/home/linuxbrew/.linuxbrew/bin:/home/linuxbrew/.linuxbrew/sbin:$HOME/.opencode/bin:$HOME/.local/bin:$HOME/bin:$PATH"
if command -v brew &>/dev/null; then echo "brew:installed"; else echo "brew:missing"; fi
if command -v opencode &>/dev/null; then echo "opencode:installed"; else echo "opencode:missing"; fi
`
		cmd := client.InstallTool(vmName, []byte(checkScript))
		out, err := cmd.Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				for _, t := range availableTools {
					if line == t.name+":installed" {
						status[t.name] = true
					} else if line == t.name+":missing" {
						status[t.name] = false
					}
				}
			}
		}
		return toolStatusMsg(status)
	}
}

func (m *ToolPickerModel) Init() tea.Cmd {
	return m.checkToolsCmd()
}

func (m *ToolPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case toolStatusMsg:
		m.status = map[string]bool(msg)
		m.loading = false

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		switch msg.String() {
		case "up", "k":
			for i := m.selected - 1; i >= 0; i-- {
				if !m.status[m.tools[i].name] {
					m.selected = i
					break
				}
			}
		case "down", "j":
			for i := m.selected + 1; i < len(m.tools); i++ {
				if !m.status[m.tools[i].name] {
					m.selected = i
					break
				}
			}
		case "enter":
			if m.status[m.tools[m.selected].name] {
				return m, nil
			}
			return m.executeTool()
		case "esc":
			action := NewActionMenuModel(m.client, m.vm, m.width, m.height)
			return m, func() tea.Msg {
				return ChangeScreenMsg{State: StateActionMenu, Screen: action}
			}
		}
	}
	return m, nil
}

func (m *ToolPickerModel) executeTool() (tea.Model, tea.Cmd) {
	t := m.tools[m.selected]

	// If this tool asks about symlinks, show the confirm dialog first
	if t.askSymlinks {
		return m.askSymlinksConfirm(t)
	}

	return m.runTool(t, t.script())
}

// askSymlinksConfirm shows a confirmation dialog asking whether to import host config.
func (m *ToolPickerModel) askSymlinksConfirm(t tool) (tea.Model, tea.Cmd) {
	client := m.client
	vm := m.vm
	width := m.width
	height := m.height

	confirm := NewConfirmModel(
		"Import host configuration via symlinks?\n\nThis will link your Mac's OpenCode config and auth\ninto the VM so you don't need to re-authenticate:\n\n  ~/.config/opencode → /Users/<you>/.config/opencode\n  ~/.local/share/opencode/auth.json → /Users/<you>/...",
		// Y — install with symlinks (default script)
		func() tea.Msg {
			installCmd := client.InstallTool(vm.Name, t.script())
			logView := NewLogViewModel(
				client, installCmd,
				fmt.Sprintf("Installing %s on %s...", t.name, vm.Name),
				width, height,
				func(exitCode int) tea.Msg {
					return reloadActionMenu(client, vm, width, height)
				},
			)
			return ChangeScreenMsg{State: StateLogView, Screen: logView}
		},
		// N — install without symlinks (prepend SKIP_SYMLINKS=1)
		func() tea.Msg {
			script := append([]byte("export SKIP_SYMLINKS=1\n"), t.script()...)
			installCmd := client.InstallTool(vm.Name, script)
			logView := NewLogViewModel(
				client, installCmd,
				fmt.Sprintf("Installing %s on %s...", t.name, vm.Name),
				width, height,
				func(exitCode int) tea.Msg {
					return reloadActionMenu(client, vm, width, height)
				},
			)
			return ChangeScreenMsg{State: StateLogView, Screen: logView}
		},
	)
	return m, func() tea.Msg {
		return ChangeScreenMsg{State: StateConfirm, Screen: confirm}
	}
}

// runTool launches the install log view for the given tool and script.
func (m *ToolPickerModel) runTool(t tool, script []byte) (tea.Model, tea.Cmd) {
	client := m.client
	vm := m.vm
	width := m.width
	height := m.height
	installCmd := client.InstallTool(vm.Name, script)
	logView := NewLogViewModel(
		client, installCmd,
		fmt.Sprintf("Installing %s on %s...", t.name, vm.Name),
		width, height,
		func(exitCode int) tea.Msg {
			return reloadActionMenu(client, vm, width, height)
		},
	)
	return m, func() tea.Msg {
		return ChangeScreenMsg{State: StateLogView, Screen: logView}
	}
}

// reloadActionMenu reloads the VM from limactl and returns to the action menu.
func reloadActionMenu(client *lima.Client, vm lima.VM, width, height int) tea.Msg {
	vms, err := client.ListVMs()
	updatedVM := vm
	if err == nil {
		for _, v := range vms {
			if v.Name == vm.Name {
				updatedVM = v
				break
			}
		}
	}
	action := NewActionMenuModel(client, updatedVM, width, height)
	return ChangeScreenMsg{State: StateActionMenu, Screen: action}
}

func (m *ToolPickerModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(fmt.Sprintf("Install Tool — %s", m.vm.Name)) + "\n\n")

	if m.loading {
		b.WriteString(MutedStyle.Render("Checking installed tools...") + "\n")
		return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	}

	for i, t := range m.tools {
		installed, checked := m.status[t.name]
		var badge string
		if !checked {
			badge = MutedStyle.Render("[ ?             ]")
		} else if installed {
			badge = RunningBadgeStyle.Render("[✓ installed    ]")
		} else {
			badge = MutedStyle.Render("[ not installed ]")
		}

		line := fmt.Sprintf("%-20s %s  %s", t.name, badge, MutedStyle.Render(t.description))
		if installed && checked {
			b.WriteString(MutedStyle.Render("   "+line) + "\n")
		} else if i == m.selected {
			b.WriteString(SelectedRowStyle.Render(" ▶ "+line) + "\n")
		} else {
			b.WriteString(NormalRowStyle.Render("   "+line) + "\n")
		}
	}

	b.WriteString("\n" + StatusBarStyle.Width(m.width).Render(
		KeyHintStyle.Render("↑↓")+" select  "+
			KeyHintStyle.Render("enter")+" install  "+
			KeyHintStyle.Render("esc")+" back",
	))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
