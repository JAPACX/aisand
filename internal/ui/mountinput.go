package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/japacx/aisand/internal/lima"
)

// MountInputModel handles adding a new mount to a VM.
type MountInputModel struct {
	client *lima.Client
	vm     lima.VM
	input  textinput.Model
	errMsg string
	width  int
	height int
}

// NewMountInputModel creates a new MountInputModel.
func NewMountInputModel(client *lima.Client, vm lima.VM, width, height int) *MountInputModel {
	ti := textinput.New()
	ti.Placeholder = "/path/to/directory"
	ti.Focus()
	ti.CharLimit = 256

	return &MountInputModel{
		client: client,
		vm:     vm,
		input:  ti,
		width:  width,
		height: height,
	}
}

func (m *MountInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *MountInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m.handleSubmit()
		case tea.KeyEsc:
			action := NewActionMenuModel(m.client, m.vm, m.width, m.height)
			return m, func() tea.Msg {
				return ChangeScreenMsg{State: StateActionMenu, Screen: action}
			}
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *MountInputModel) handleSubmit() (tea.Model, tea.Cmd) {
	path := strings.TrimSpace(m.input.Value())
	if path == "" {
		m.errMsg = "Path cannot be empty"
		return m, nil
	}

	// Validate path exists on host
	if _, err := os.Stat(path); os.IsNotExist(err) {
		m.errMsg = "Path does not exist on host"
		return m, nil
	}

	// Validate not already mounted
	for _, mount := range m.vm.Mounts {
		if mount.Location == path {
			m.errMsg = "Path is already mounted"
			return m, nil
		}
	}

	vmWasRunning := m.vm.Status == "Running"
	vmName := m.vm.Name

	if vmWasRunning {
		// Show restart warning
		confirm := NewConfirmModel(
			fmt.Sprintf("Adding mount %q requires a VM restart. Continue?", path),
			func() tea.Msg {
				_ = m.client.StopVM(vmName)
				_ = m.client.AddMount(vmName, path)
				startCmd := m.client.StartVM(vmName)
				logView := NewLogViewModel(
					m.client, startCmd,
					fmt.Sprintf("Restarting %s after mount...", vmName),
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

	// VM is stopped — just add the mount
	_ = m.client.AddMount(vmName, path)
	main := NewMainModel(m.client, m.width, m.height)
	return m, func() tea.Msg {
		return ChangeScreenMsg{State: StateMain, Screen: main}
	}
}

func (m *MountInputModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(fmt.Sprintf("Add Mount — %s", m.vm.Name)) + "\n\n")
	b.WriteString("Enter the host directory path to mount:\n\n")
	b.WriteString(m.input.View() + "\n")

	if m.errMsg != "" {
		b.WriteString(ErrorStyle.Render(m.errMsg) + "\n")
	}

	b.WriteString("\n" + StatusBarStyle.Width(m.width).Render(
		KeyHintStyle.Render("enter")+" confirm  "+
			KeyHintStyle.Render("esc")+" cancel",
	))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
