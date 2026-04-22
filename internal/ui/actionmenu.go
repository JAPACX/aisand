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
	stopped := vm.Status != "Running"
	items := []actionItem{
		{label: "Shell", disabled: stopped},
		{label: "Start", disabled: vm.Status == "Running"},
		{label: "Stop", disabled: vm.Status == "Stopped"},
		{label: "Delete"},
		{label: "Install tool", disabled: stopped},
		{label: "Mount", disabled: stopped},
		{label: "Unmount", disabled: stopped},
		{label: "Mounts", disabled: stopped},
	}
	m := &ActionMenuModel{
		client: client,
		vm:     vm,
		items:  items,
		width:  width,
		height: height,
	}
	// Ensure initial selection is on an enabled item
	for i, item := range items {
		if !item.disabled {
			m.selected = i
			break
		}
	}
	return m
}

// backToAction returns a tea.Cmd that reloads the VM from limactl and transitions
// back to the ActionMenu with fresh state. Used after any non-destructive operation.
func (m *ActionMenuModel) backToAction() tea.Cmd {
	client := m.client
	vmName := m.vm.Name
	width := m.width
	height := m.height
	return func() tea.Msg {
		// Reload VM list to get fresh status
		vms, err := client.ListVMs()
		vm := lima.VM{Name: vmName} // fallback if reload fails
		if err == nil {
			for _, v := range vms {
				if v.Name == vmName {
					vm = v
					break
				}
			}
		}
		action := NewActionMenuModel(client, vm, width, height)
		return ChangeScreenMsg{State: StateActionMenu, Screen: action}
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
		m.selected = m.prevEnabled(m.selected)
	case "down", "j":
		m.selected = m.nextEnabled(m.selected)
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

// nextEnabled returns the index of the next enabled item after current, or current if none.
func (m *ActionMenuModel) nextEnabled(current int) int {
	for i := current + 1; i < len(m.items); i++ {
		if !m.items[i].disabled {
			return i
		}
	}
	return current // already at last enabled, stay
}

// prevEnabled returns the index of the previous enabled item before current, or current if none.
func (m *ActionMenuModel) prevEnabled(current int) int {
	for i := current - 1; i >= 0; i-- {
		if !m.items[i].disabled {
			return i
		}
	}
	return current // already at first enabled, stay
}

func (m *ActionMenuModel) executeAction() (tea.Model, tea.Cmd) {
	item := m.items[m.selected]
	if item.disabled {
		return m, nil
	}

	switch item.label {
	case "Shell":
		if m.vm.Status == "Stopped" {
			startCmd := m.client.StartVM(m.vm.Name)
			logView := NewLogViewModel(
				m.client, startCmd,
				fmt.Sprintf("Starting %s...", m.vm.Name),
				m.width, m.height,
				func(exitCode int) tea.Msg {
					if exitCode == 0 {
						shellCmd := m.client.ShellVM(m.vm.Name)
						return tea.ExecProcess(shellCmd, func(err error) tea.Msg {
							return m.backToAction()()
						})
					}
					return m.backToAction()()
				},
			)
			return m, func() tea.Msg {
				return ChangeScreenMsg{State: StateLogView, Screen: logView}
			}
		}
		shellCmd := m.client.ShellVM(m.vm.Name)
		return m, tea.ExecProcess(shellCmd, func(err error) tea.Msg {
			return m.backToAction()()
		})

	case "Start":
		startCmd := m.client.StartVM(m.vm.Name)
		logView := NewLogViewModelWithHint(
			m.client, startCmd,
			fmt.Sprintf("Starting %s...", m.vm.Name),
			"This operation may take a moment, please wait.",
			m.width, m.height,
			func(exitCode int) tea.Msg {
				return m.backToAction()()
			},
		)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateLogView, Screen: logView}
		}

	case "Stop":
		stopCmd := m.client.StopVM2(m.vm.Name)
		logView := NewLogViewModel(
			m.client, stopCmd,
			fmt.Sprintf("Stopping %s...", m.vm.Name),
			m.width, m.height,
			func(exitCode int) tea.Msg {
				return m.backToAction()()
			},
		)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateLogView, Screen: logView}
		}

	case "Delete":
		confirm := NewConfirmModel(
			fmt.Sprintf("Delete VM %q? This cannot be undone.", m.vm.Name),
			func() tea.Msg {
				deleteCmd := m.client.DeleteVMCmd(m.vm.Name)
				logView := NewLogViewModel(
					m.client, deleteCmd,
					fmt.Sprintf("Deleting %s...", m.vm.Name),
					m.width, m.height,
					func(exitCode int) tea.Msg {
						main := NewMainModel(m.client, m.width, m.height)
						return ChangeScreenMsg{State: StateMain, Screen: main}
					},
				)
				return ChangeScreenMsg{State: StateLogView, Screen: logView}
			},
			func() tea.Msg {
				return m.backToAction()()
			},
		)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateConfirm, Screen: confirm}
		}

	case "Install tool":
		picker := NewToolPickerModel(m.client, m.vm, m.width, m.height)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateToolPicker, Screen: picker}
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
