package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/japacx/aisand/internal/embed"
	"github.com/japacx/aisand/internal/lima"
)

type createVMStep int

const (
	stepName createVMStep = iota
	stepCPUs
	stepRAM
	stepDisk
	stepMounts
	stepSummary
	stepCreating
)

// CreateVMModel is the multi-step VM creation form.
type CreateVMModel struct {
	client      *lima.Client
	existingVMs []lima.VM
	step        createVMStep

	// Name step
	nameInput textinput.Model
	nameError string

	// CPUs step
	cpuOptions []int
	cpuIdx     int

	// RAM step
	ramOptions []int
	ramIdx     int

	// Disk step
	diskOptions []int
	diskIdx     int

	// Mounts step
	mountInput  textinput.Model
	mounts      []string
	mountError  string

	// Summary
	confirmed bool

	width  int
	height int
}

// NewCreateVMModel creates a new CreateVMModel.
func NewCreateVMModel(client *lima.Client, existingVMs []lima.VM, width, height int) *CreateVMModel {
	hostCPUs := client.GetHostCPUs()
	maxCPUs := hostCPUs / 2
	if maxCPUs > 4 {
		maxCPUs = 4
	}
	if maxCPUs < 1 {
		maxCPUs = 1
	}
	cpuOptions := make([]int, maxCPUs)
	for i := range cpuOptions {
		cpuOptions[i] = i + 1
	}

	nameInput := textinput.New()
	nameInput.Placeholder = "my-vm"
	nameInput.Focus()
	nameInput.CharLimit = 64

	mountInput := textinput.New()
	mountInput.Placeholder = "/path/to/dir (empty to finish)"
	mountInput.CharLimit = 256

	return &CreateVMModel{
		client:      client,
		existingVMs: existingVMs,
		step:        stepName,
		nameInput:   nameInput,
		cpuOptions:  cpuOptions,
		cpuIdx:      len(cpuOptions) - 1,
		ramOptions:  []int{1, 2, 4, 8},
		ramIdx:      2, // default 4 GB
		diskOptions: []int{20, 40, 60, 80, 100},
		diskIdx:     2, // default 60 GB
		mountInput:  mountInput,
		width:       width,
		height:      height,
	}
}

func (m *CreateVMModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *CreateVMModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Update active text input
	var cmd tea.Cmd
	switch m.step {
	case stepName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case stepMounts:
		m.mountInput, cmd = m.mountInput.Update(msg)
	}
	return m, cmd
}

func (m *CreateVMModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.step {
	case stepName:
		switch msg.Type {
		case tea.KeyEnter:
			name := strings.TrimSpace(m.nameInput.Value())
			if name == "" {
				m.nameError = "Name cannot be empty"
				return m, nil
			}
			if strings.Contains(name, " ") {
				m.nameError = "Name cannot contain spaces"
				return m, nil
			}
			for _, vm := range m.existingVMs {
				if vm.Name == name {
					m.nameError = fmt.Sprintf("VM %q already exists", name)
					return m, nil
				}
			}
			m.nameError = ""
			m.step = stepCPUs
		case tea.KeyEsc:
			main := NewMainModel(m.client, m.width, m.height)
			return m, func() tea.Msg {
				return ChangeScreenMsg{State: StateMain, Screen: main}
			}
		}

	case stepCPUs:
		switch msg.String() {
		case "up", "k":
			if m.cpuIdx > 0 {
				m.cpuIdx--
			}
		case "down", "j":
			if m.cpuIdx < len(m.cpuOptions)-1 {
				m.cpuIdx++
			}
		case "enter":
			m.step = stepRAM
		case "esc":
			m.step = stepName
		}

	case stepRAM:
		switch msg.String() {
		case "up", "k":
			if m.ramIdx > 0 {
				m.ramIdx--
			}
		case "down", "j":
			if m.ramIdx < len(m.ramOptions)-1 {
				m.ramIdx++
			}
		case "enter":
			m.step = stepDisk
		case "esc":
			m.step = stepCPUs
		}

	case stepDisk:
		switch msg.String() {
		case "up", "k":
			if m.diskIdx > 0 {
				m.diskIdx--
			}
		case "down", "j":
			if m.diskIdx < len(m.diskOptions)-1 {
				m.diskIdx++
			}
		case "enter":
			m.step = stepMounts
			m.mountInput.SetValue("")
			m.mountInput.Focus()
		case "esc":
			m.step = stepRAM
		}

	case stepMounts:
		switch msg.Type {
		case tea.KeyEnter:
			path := strings.TrimSpace(m.mountInput.Value())
			if path == "" {
				// Empty = done with mounts
				m.step = stepSummary
				return m, nil
			}
			// Validate path exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				m.mountError = fmt.Sprintf("Path does not exist: %s", path)
				return m, nil
			}
			// Check for duplicates
			for _, existing := range m.mounts {
				if existing == path {
					m.mountError = fmt.Sprintf("Path already added: %s", path)
					return m, nil
				}
			}
			m.mounts = append(m.mounts, path)
			m.mountError = ""
			m.mountInput.SetValue("")
		case tea.KeyEsc:
			if len(m.mounts) > 0 {
				// Remove last mount
				m.mounts = m.mounts[:len(m.mounts)-1]
			} else {
				m.step = stepDisk
			}
		}

	case stepSummary:
		switch msg.String() {
		case "y", "Y":
			return m.startCreation()
		case "n", "N", "esc":
			main := NewMainModel(m.client, m.width, m.height)
			return m, func() tea.Msg {
				return ChangeScreenMsg{State: StateMain, Screen: main}
			}
		}
	}

	// Update text inputs
	var cmd tea.Cmd
	switch m.step {
	case stepName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case stepMounts:
		m.mountInput, cmd = m.mountInput.Update(msg)
	}
	return m, cmd
}

func (m *CreateVMModel) startCreation() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.nameInput.Value())
	cpus := m.cpuOptions[m.cpuIdx]
	ram := m.ramOptions[m.ramIdx]
	disk := m.diskOptions[m.diskIdx]

	// Write embedded YAML to temp file
	tmpFile, err := os.CreateTemp("", "aisand-template-*.yaml")
	if err != nil {
		m.nameError = "Failed to create temp file: " + err.Error()
		return m, nil
	}
	if _, err := tmpFile.Write(embed.AgentYAML); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		m.nameError = "Failed to write template: " + err.Error()
		return m, nil
	}
	tmpFile.Close()
	templatePath := tmpFile.Name()

	createCmd := m.client.CreateVM(name, cpus, ram, disk, m.mounts, templatePath)
	logView := NewLogViewModel(
		m.client, createCmd,
		fmt.Sprintf("Creating VM %q...", name),
		m.width, m.height,
		func(exitCode int) tea.Msg {
			// Clean up temp file
			os.Remove(templatePath)
			main := NewMainModel(m.client, m.width, m.height)
			return ChangeScreenMsg{State: StateMain, Screen: main}
		},
	)

	return m, func() tea.Msg {
		return ChangeScreenMsg{State: StateLogView, Screen: logView}
	}
}

func (m *CreateVMModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Create New VM") + "\n\n")

	switch m.step {
	case stepName:
		b.WriteString("Step 1/6: VM Name\n\n")
		b.WriteString(m.nameInput.View() + "\n")
		if m.nameError != "" {
			b.WriteString(ErrorStyle.Render(m.nameError) + "\n")
		}
		b.WriteString("\n" + StatusBarStyle.Width(m.width).Render(
			KeyHintStyle.Render("enter")+" confirm  "+KeyHintStyle.Render("esc")+" cancel",
		))

	case stepCPUs:
		b.WriteString(fmt.Sprintf("Step 2/6: CPUs (VM: %s)\n\n", m.nameInput.Value()))
		for i, v := range m.cpuOptions {
			label := fmt.Sprintf("%d CPUs", v)
			if i == m.cpuIdx {
				b.WriteString(SelectedRowStyle.Render(" ▶ "+label) + "\n")
			} else {
				b.WriteString(NormalRowStyle.Render("   "+label) + "\n")
			}
		}
		b.WriteString("\n" + StatusBarStyle.Width(m.width).Render(
			KeyHintStyle.Render("↑↓")+" select  "+KeyHintStyle.Render("enter")+" confirm  "+KeyHintStyle.Render("esc")+" back",
		))

	case stepRAM:
		b.WriteString(fmt.Sprintf("Step 3/6: RAM (CPUs: %d)\n\n", m.cpuOptions[m.cpuIdx]))
		for i, v := range m.ramOptions {
			label := fmt.Sprintf("%d GB", v)
			if i == m.ramIdx {
				b.WriteString(SelectedRowStyle.Render(" ▶ "+label) + "\n")
			} else {
				b.WriteString(NormalRowStyle.Render("   "+label) + "\n")
			}
		}
		b.WriteString("\n" + StatusBarStyle.Width(m.width).Render(
			KeyHintStyle.Render("↑↓")+" select  "+KeyHintStyle.Render("enter")+" confirm  "+KeyHintStyle.Render("esc")+" back",
		))

	case stepDisk:
		b.WriteString(fmt.Sprintf("Step 4/6: Disk (RAM: %d GB)\n\n", m.ramOptions[m.ramIdx]))
		for i, v := range m.diskOptions {
			label := fmt.Sprintf("%d GB", v)
			if i == m.diskIdx {
				b.WriteString(SelectedRowStyle.Render(" ▶ "+label) + "\n")
			} else {
				b.WriteString(NormalRowStyle.Render("   "+label) + "\n")
			}
		}
		b.WriteString("\n" + StatusBarStyle.Width(m.width).Render(
			KeyHintStyle.Render("↑↓")+" select  "+KeyHintStyle.Render("enter")+" confirm  "+KeyHintStyle.Render("esc")+" back",
		))

	case stepMounts:
		b.WriteString("Step 5/6: Host Mounts (optional)\n\n")
		if len(m.mounts) > 0 {
			b.WriteString("Added mounts:\n")
			for _, mp := range m.mounts {
				b.WriteString(NormalRowStyle.Render("  + "+mp) + "\n")
			}
			b.WriteString("\n")
		}
		b.WriteString(m.mountInput.View() + "\n")
		if m.mountError != "" {
			b.WriteString(ErrorStyle.Render(m.mountError) + "\n")
		}
		b.WriteString(MutedStyle.Render("(leave empty and press enter to continue)") + "\n")
		b.WriteString("\n" + StatusBarStyle.Width(m.width).Render(
			KeyHintStyle.Render("enter")+" add/continue  "+KeyHintStyle.Render("esc")+" remove last/back",
		))

	case stepSummary:
		b.WriteString("Step 6/6: Summary\n\n")
		b.WriteString(DetailLabelStyle.Render("Name:") + DetailValueStyle.Render(m.nameInput.Value()) + "\n")
		b.WriteString(DetailLabelStyle.Render("CPUs:") + DetailValueStyle.Render(fmt.Sprintf("%d", m.cpuOptions[m.cpuIdx])) + "\n")
		b.WriteString(DetailLabelStyle.Render("RAM:") + DetailValueStyle.Render(fmt.Sprintf("%d GB", m.ramOptions[m.ramIdx])) + "\n")
		b.WriteString(DetailLabelStyle.Render("Disk:") + DetailValueStyle.Render(fmt.Sprintf("%d GB", m.diskOptions[m.diskIdx])) + "\n")
		if len(m.mounts) > 0 {
			b.WriteString(DetailLabelStyle.Render("Mounts:") + "\n")
			for _, mp := range m.mounts {
				b.WriteString("  " + NormalRowStyle.Render(mp) + "\n")
			}
		} else {
			b.WriteString(DetailLabelStyle.Render("Mounts:") + MutedStyle.Render("none") + "\n")
		}
		b.WriteString("\n" + KeyHintStyle.Render("y") + " to create, " + KeyHintStyle.Render("n") + " to cancel\n")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
