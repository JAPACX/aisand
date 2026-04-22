package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/japacx/aisand/internal/lima"
)

type actionItem struct {
	label    string
	disabled bool
}

// ActionMenuModel is the contextual action menu for a selected VM.
type ActionMenuModel struct {
	client   *lima.Client
	vm       lima.VM
	items    []actionItem
	selected int
	width    int
	height   int
}

// NewActionMenuModel creates a new ActionMenuModel for the given VM.
func NewActionMenuModel(client *lima.Client, vm lima.VM, width, height int) *ActionMenuModel {
	items := []actionItem{
		{label: "Shell"},
		{label: "Start", disabled: vm.Status == "Running"},
		{label: "Stop", disabled: vm.Status == "Stopped"},
		{label: "Delete"},
		{label: "Install tool"},
		{label: "Mount"},
		{label: "Unmount"},
		{label: "Mounts"},
	}
	return &ActionMenuModel{
		client: client,
		vm:     vm,
		items:  items,
		width:  width,
		height: height,
	}
}

func (m *ActionMenuModel) Init() tea.Cmd { return nil }

func (m *ActionMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *ActionMenuModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.selected = (m.selected - 1 + len(m.items)) % len(m.items)
	case "down", "j":
		m.selected = (m.selected + 1) % len(m.items)
	case "enter":
		return m.executeAction()
	case "esc":
		main := NewMainModel(m.client, m.width, m.height)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateMain, Screen: main}
		}
	}
	return m, nil
}

func (m *ActionMenuModel) executeAction() (tea.Model, tea.Cmd) {
	item := m.items[m.selected]
	if item.disabled {
		return m, nil
	}

	switch item.label {
	case "Shell":
		// If VM is stopped, start it first then shell
		if m.vm.Status == "Stopped" {
			startCmd := m.client.StartVM(m.vm.Name)
			logView := NewLogViewModel(
				m.client, startCmd,
				fmt.Sprintf("Starting %s...", m.vm.Name),
				m.width, m.height,
				func(exitCode int) tea.Msg {
					if exitCode == 0 {
						// After start, open shell
						shellCmd := m.client.ShellVM(m.vm.Name)
						return tea.ExecProcess(shellCmd, func(err error) tea.Msg {
							main := NewMainModel(m.client, m.width, m.height)
							return ChangeScreenMsg{State: StateMain, Screen: main}
						})
					}
					main := NewMainModel(m.client, m.width, m.height)
					return ChangeScreenMsg{State: StateMain, Screen: main}
				},
			)
			return m, func() tea.Msg {
				return ChangeScreenMsg{State: StateLogView, Screen: logView}
			}
		}
		// VM is running — open shell directly
		shellCmd := m.client.ShellVM(m.vm.Name)
		return m, tea.ExecProcess(shellCmd, func(err error) tea.Msg {
			main := NewMainModel(m.client, m.width, m.height)
			return ChangeScreenMsg{State: StateMain, Screen: main}
		})

	case "Start":
		startCmd := m.client.StartVM(m.vm.Name)
		logView := NewLogViewModel(
			m.client, startCmd,
			fmt.Sprintf("Starting %s...", m.vm.Name),
			m.width, m.height,
			func(exitCode int) tea.Msg {
				main := NewMainModel(m.client, m.width, m.height)
				return ChangeScreenMsg{State: StateMain, Screen: main}
			},
		)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateLogView, Screen: logView}
		}

	case "Stop":
		return m, func() tea.Msg {
			_ = m.client.StopVM(m.vm.Name)
			main := NewMainModel(m.client, m.width, m.height)
			return ChangeScreenMsg{State: StateMain, Screen: main}
		}

	case "Delete":
		confirm := NewConfirmModel(
			fmt.Sprintf("Delete VM %q? This cannot be undone.", m.vm.Name),
			func() tea.Msg {
				_ = m.client.DeleteVM(m.vm.Name)
				main := NewMainModel(m.client, m.width, m.height)
				return ChangeScreenMsg{State: StateMain, Screen: main}
			},
			func() tea.Msg {
				main := NewMainModel(m.client, m.width, m.height)
				return ChangeScreenMsg{State: StateMain, Screen: main}
			},
		)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateConfirm, Screen: confirm}
		}

	case "Install tool":
		// Get the embedded opencode script from the embed package
		installCmd := m.client.InstallTool(m.vm.Name, getOpenCodeScript())
		logView := NewLogViewModel(
			m.client, installCmd,
			fmt.Sprintf("Installing opencode on %s...", m.vm.Name),
			m.width, m.height,
			func(exitCode int) tea.Msg {
				main := NewMainModel(m.client, m.width, m.height)
				return ChangeScreenMsg{State: StateMain, Screen: main}
			},
		)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateLogView, Screen: logView}
		}

	case "Mount":
		mountInput := NewMountInputModel(m.client, m.vm, m.width, m.height)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateLogView, Screen: mountInput}
		}

	case "Unmount":
		picker := NewUnmountPickerModel(m.client, m.vm, m.width, m.height)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateUnmountPicker, Screen: picker}
		}

	case "Mounts":
		mountsList := NewMountsListModel(m.client, m.vm, m.width, m.height)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateMountsList, Screen: mountsList}
		}
	}

	return m, nil
}

func (m *ActionMenuModel) View() string {
	var b strings.Builder

	title := TitleStyle.Render(fmt.Sprintf("Actions — %s (%s)", m.vm.Name, m.vm.Status))
	b.WriteString(title + "\n\n")

	for i, item := range m.items {
		label := fmt.Sprintf("%d. %s", i+1, item.label)
		if item.disabled {
			b.WriteString(MutedStyle.Render("   "+label) + "\n")
		} else if i == m.selected {
			b.WriteString(SelectedRowStyle.Render(" ▶ "+label) + "\n")
		} else {
			b.WriteString(NormalRowStyle.Render("   "+label) + "\n")
		}
	}

	b.WriteString("\n" + StatusBarStyle.Width(m.width).Render(
		KeyHintStyle.Render("↑↓")+" select  "+
			KeyHintStyle.Render("enter")+" confirm  "+
			KeyHintStyle.Render("esc")+" back",
	))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
