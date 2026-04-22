package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/japacx/aisand/internal/lima"
)

// MountsListModel displays all mounts for a VM.
type MountsListModel struct {
	client *lima.Client
	vm     lima.VM
	width  int
	height int
}

// NewMountsListModel creates a new MountsListModel.
func NewMountsListModel(client *lima.Client, vm lima.VM, width, height int) *MountsListModel {
	return &MountsListModel{client: client, vm: vm, width: width, height: height}
}

func (m *MountsListModel) Init() tea.Cmd { return nil }

func (m *MountsListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "esc" {
			action := NewActionMenuModel(m.client, m.vm, m.width, m.height)
			return m, func() tea.Msg {
				return ChangeScreenMsg{State: StateActionMenu, Screen: action}
			}
		}
	}
	return m, nil
}

func (m *MountsListModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(fmt.Sprintf("Mounts — %s", m.vm.Name)) + "\n\n")

	if len(m.vm.Mounts) == 0 {
		b.WriteString(MutedStyle.Render("No additional mounts configured.") + "\n")
	} else {
		for _, mount := range m.vm.Mounts {
			mode := "read-only"
			if mount.Writable {
				mode = "read-write"
			}
			b.WriteString(NormalRowStyle.Render(fmt.Sprintf("  %s (%s)", mount.Location, mode)) + "\n")
		}
	}

	b.WriteString("\n" + StatusBarStyle.Width(m.width).Render(
		KeyHintStyle.Render("esc")+" back",
	))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

// UnmountPickerModel lets the user select a mount to remove.
type UnmountPickerModel struct {
	client   *lima.Client
	vm       lima.VM
	selected int
	width    int
	height   int
}

// NewUnmountPickerModel creates a new UnmountPickerModel.
func NewUnmountPickerModel(client *lima.Client, vm lima.VM, width, height int) *UnmountPickerModel {
	return &UnmountPickerModel{client: client, vm: vm, width: width, height: height}
}

func (m *UnmountPickerModel) Init() tea.Cmd { return nil }

func (m *UnmountPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *UnmountPickerModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	mounts := m.vm.Mounts
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(mounts)-1 {
			m.selected++
		}
	case "enter":
		if len(mounts) == 0 {
			return m, nil
		}
		selectedMount := mounts[m.selected]
		// Build remaining mounts (all except selected)
		remaining := make([]lima.Mount, 0, len(mounts)-1)
		for i, mount := range mounts {
			if i != m.selected {
				remaining = append(remaining, mount)
			}
		}

		vmWasRunning := m.vm.Status == "Running"
		vmName := m.vm.Name

		if vmWasRunning {
			// Show restart warning
			confirm := NewConfirmModel(
				fmt.Sprintf("Removing mount %q requires a VM restart. Continue?", selectedMount.Location),
				func() tea.Msg {
					// Stop, edit, start
					_ = m.client.StopVM(vmName)
					_ = m.client.RemoveMount(vmName, remaining)
					startCmd := m.client.StartVM(vmName)
					logView := NewLogViewModel(
						m.client, startCmd,
						fmt.Sprintf("Restarting %s after unmount...", vmName),
						m.width, m.height,
						func(exitCode int) tea.Msg {
							main := NewMainModel(m.client, m.width, m.height)
							return ChangeScreenMsg{State: StateMain, Screen: main}
						},
					)
					return ChangeScreenMsg{State: StateLogView, Screen: logView}
				},
				func() tea.Msg {
					action := NewActionMenuModel(m.client, m.vm, m.width, m.height)
					return ChangeScreenMsg{State: StateActionMenu, Screen: action}
				},
			)
			return m, func() tea.Msg {
				return ChangeScreenMsg{State: StateConfirm, Screen: confirm}
			}
		}

		// VM is stopped — just remove the mount
		_ = m.client.RemoveMount(vmName, remaining)
		main := NewMainModel(m.client, m.width, m.height)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateMain, Screen: main}
		}

	case "esc":
		action := NewActionMenuModel(m.client, m.vm, m.width, m.height)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateActionMenu, Screen: action}
		}
	}
	return m, nil
}

func (m *UnmountPickerModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(fmt.Sprintf("Remove Mount — %s", m.vm.Name)) + "\n\n")

	if len(m.vm.Mounts) == 0 {
		b.WriteString(MutedStyle.Render("No mounts to remove.") + "\n")
	} else {
		for i, mount := range m.vm.Mounts {
			mode := "ro"
			if mount.Writable {
				mode = "rw"
			}
			label := fmt.Sprintf("%s (%s)", mount.Location, mode)
			if i == m.selected {
				b.WriteString(SelectedRowStyle.Render(" ▶ "+label) + "\n")
			} else {
				b.WriteString(NormalRowStyle.Render("   "+label) + "\n")
			}
		}
	}

	b.WriteString("\n" + StatusBarStyle.Width(m.width).Render(
		KeyHintStyle.Render("↑↓")+" select  "+
			KeyHintStyle.Render("enter")+" remove  "+
			KeyHintStyle.Render("esc")+" back",
	))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
