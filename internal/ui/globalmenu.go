package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/japacx/aisand/internal/lima"
)

// GlobalMenuModel is the global operations menu.
type GlobalMenuModel struct {
	client   *lima.Client
	vms      []lima.VM
	items    []string
	selected int
	width    int
	height   int
	message  string // informational message (e.g. "No VMs running")
}

// NewGlobalMenuModel creates a new GlobalMenuModel.
func NewGlobalMenuModel(client *lima.Client, vms []lima.VM, width, height int) *GlobalMenuModel {
	return &GlobalMenuModel{
		client: client,
		vms:    vms,
		items:  []string{"New VM", "Stop all VMs", "Host setup", "Refresh"},
		width:  width,
		height: height,
	}
}

func (m *GlobalMenuModel) Init() tea.Cmd { return nil }

func (m *GlobalMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *GlobalMenuModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.selected = (m.selected - 1 + len(m.items)) % len(m.items)
		m.message = ""
	case "down", "j":
		m.selected = (m.selected + 1) % len(m.items)
		m.message = ""
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

func (m *GlobalMenuModel) executeAction() (tea.Model, tea.Cmd) {
	switch m.items[m.selected] {
	case "New VM":
		createVM := NewCreateVMModel(m.client, m.vms, m.width, m.height)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateCreateVM, Screen: createVM}
		}

	case "Stop all VMs":
		// Check if any VMs are running
		running := 0
		for _, vm := range m.vms {
			if vm.Status == "Running" {
				running++
			}
		}
		if running == 0 {
			m.message = "No VMs are running."
			return m, nil
		}
		confirm := NewConfirmModel(
			fmt.Sprintf("Stop all %d running VM(s)?", running),
			func() tea.Msg {
				_ = m.client.StopAllVMs(m.vms)
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

	case "Host setup":
		setup := NewSetupModel(m.client)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateSetup, Screen: setup}
		}

	case "Refresh":
		return m, func() tea.Msg {
			vms, _ := m.client.ListVMs()
			main := NewMainModel(m.client, m.width, m.height)
			main.vms = vms
			return ChangeScreenMsg{State: StateMain, Screen: main}
		}
	}

	return m, nil
}

func (m *GlobalMenuModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Global Menu") + "\n\n")

	for i, item := range m.items {
		label := fmt.Sprintf("%d. %s", i+1, item)
		if i == m.selected {
			b.WriteString(SelectedRowStyle.Render(" ▶ "+label) + "\n")
		} else {
			b.WriteString(NormalRowStyle.Render("   "+label) + "\n")
		}
	}

	if m.message != "" {
		b.WriteString("\n" + MutedStyle.Render(m.message) + "\n")
	}

	b.WriteString("\n" + StatusBarStyle.Width(m.width).Render(
		KeyHintStyle.Render("↑↓")+" select  "+
			KeyHintStyle.Render("enter")+" confirm  "+
			KeyHintStyle.Render("esc")+" back",
	))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
