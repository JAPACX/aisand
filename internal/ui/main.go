package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/japacx/aisand/internal/lima"
)

// refreshVMsMsg is sent when the VM list should be reloaded.
type refreshVMsMsg struct{}

// vmListLoadedMsg carries the result of a VM list reload.
type vmListLoadedMsg struct {
	vms []lima.VM
	err error
}

// MainModel is the split-panel main screen.
type MainModel struct {
	client   *lima.Client
	vms      []lima.VM
	selected int
	err      error
	width    int
	height   int
}

// NewMainModel creates a new MainModel.
func NewMainModel(client *lima.Client, width, height int) *MainModel {
	return &MainModel{
		client: client,
		width:  width,
		height: height,
	}
}

// loadVMs is a tea.Cmd that fetches the VM list.
func (m *MainModel) loadVMs() tea.Cmd {
	return func() tea.Msg {
		vms, err := m.client.ListVMs()
		return vmListLoadedMsg{vms: vms, err: err}
	}
}

func (m *MainModel) Init() tea.Cmd {
	return m.loadVMs()
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case vmListLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.vms = msg.vms
			m.err = nil
			if m.selected >= len(m.vms) {
				m.selected = max(0, len(m.vms)-1)
			}
		}

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *MainModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(m.vms)-1 {
			m.selected++
		}
	case "enter":
		if len(m.vms) > 0 {
			vm := m.vms[m.selected]
			action := NewActionMenuModel(m.client, vm, m.width, m.height)
			return m, func() tea.Msg {
				return ChangeScreenMsg{State: StateActionMenu, Screen: action}
			}
		}
	case "g":
		global := NewGlobalMenuModel(m.client, m.vms, m.width, m.height)
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateGlobalMenu, Screen: global}
		}
	case "r":
		return m, m.loadVMs()
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m *MainModel) View() string {
	if m.width == 0 {
		return TitleStyle.Render("aisand") + "\n\n" + MutedStyle.Render("Loading...")
	}

	leftWidth := int(float64(m.width) * LeftPanelRatio)
	rightWidth := m.width - leftWidth - 4 // account for borders

	leftContent := m.renderLeft(leftWidth)
	rightContent := m.renderRight(rightWidth)

	leftPanel := PanelStyle.Width(leftWidth).Height(m.height - StatusBarHeight - 4).Render(leftContent)
	rightPanel := PanelStyle.Width(rightWidth).Height(m.height - StatusBarHeight - 4).Render(rightContent)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, panels, statusBar)
}

func (m *MainModel) renderLeft(width int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("VMs") + "\n\n")

	if m.err != nil {
		b.WriteString(ErrorStyle.Render("Error: "+m.err.Error()) + "\n")
		return b.String()
	}

	if len(m.vms) == 0 {
		b.WriteString(MutedStyle.Render("No VMs found.") + "\n")
		b.WriteString(MutedStyle.Render("Press g → New VM") + "\n")
		return b.String()
	}

	for i, vm := range m.vms {
		name := vm.Name
		badge := StatusBadge(vm.Status)
		line := fmt.Sprintf("%-*s %s", width-12, name, badge)
		if i == m.selected {
			b.WriteString(SelectedRowStyle.Render(line) + "\n")
		} else {
			b.WriteString(NormalRowStyle.Render(line) + "\n")
		}
	}
	return b.String()
}

func (m *MainModel) renderRight(width int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("VM Detail") + "\n\n")

	if len(m.vms) == 0 {
		b.WriteString(MutedStyle.Render("No VM selected.") + "\n")
		b.WriteString(MutedStyle.Render("Press g → New VM to create one.") + "\n")
		return b.String()
	}

	vm := m.vms[m.selected]

	b.WriteString(DetailLabelStyle.Render("Name:") + DetailValueStyle.Render(vm.Name) + "\n")
	b.WriteString(DetailLabelStyle.Render("Status:") + StatusBadge(vm.Status) + "\n")
	b.WriteString(DetailLabelStyle.Render("CPUs:") + DetailValueStyle.Render(fmt.Sprintf("%d", vm.CPUs)) + "\n")
	b.WriteString(DetailLabelStyle.Render("RAM:") + DetailValueStyle.Render(fmt.Sprintf("%d MiB", vm.Memory/1024/1024)) + "\n")
	b.WriteString(DetailLabelStyle.Render("Disk:") + DetailValueStyle.Render(fmt.Sprintf("%d GiB", vm.Disk/1024/1024/1024)) + "\n")

	b.WriteString("\n" + TitleStyle.Render("Mounts:") + "\n")
	if len(vm.Mounts) == 0 {
		b.WriteString(MutedStyle.Render("  No additional mounts") + "\n")
	} else {
		for _, mount := range vm.Mounts {
			mode := "ro"
			if mount.Writable {
				mode = "rw"
			}
			b.WriteString(NormalRowStyle.Render(fmt.Sprintf("  %s (%s)", mount.Location, mode)) + "\n")
		}
	}

	return b.String()
}

func (m *MainModel) renderStatusBar() string {
	keys := []string{
		KeyHintStyle.Render("↑↓") + " navigate",
		KeyHintStyle.Render("enter") + " actions",
		KeyHintStyle.Render("g") + " global",
		KeyHintStyle.Render("r") + " refresh",
		KeyHintStyle.Render("q") + " quit",
	}
	return StatusBarStyle.Width(m.width).Render(strings.Join(keys, "  "))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
