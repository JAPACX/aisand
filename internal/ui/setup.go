package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/japacx/aisand/internal/lima"
)

type setupSubState int

const (
	setupStateCheck setupSubState = iota
	setupStateConfirm
	setupStateInstalling
	setupStateConfigCPUs
	setupStateConfigRAM
	setupStateConfigDisk
	setupStateDone
)

// SetupModel is the onboarding screen shown when limactl is not installed.
type SetupModel struct {
	client       *lima.Client
	subState     setupSubState
	brewMissing  bool
	limaMissing  bool
	logLines     []string
	installDone  bool
	installError string

	// Resource config selections
	cpuOptions  []int
	ramOptions  []int
	diskOptions []int
	cpuIdx      int
	ramIdx      int
	diskIdx     int

	width  int
	height int
}

// NewSetupModel creates a new SetupModel.
func NewSetupModel(client *lima.Client) *SetupModel {
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

	return &SetupModel{
		client:      client,
		brewMissing: !client.IsBrewInstalled(),
		limaMissing: !client.IsLimactlInstalled(),
		cpuOptions:  cpuOptions,
		ramOptions:  []int{1, 2, 4, 8},
		diskOptions: []int{20, 40, 60, 80, 100},
		cpuIdx:      len(cpuOptions) - 1, // default: max
		ramIdx:      2,                   // default: 4 GB
		diskIdx:     2,                   // default: 60 GB
	}
}

// logLineMsg carries a new log line from a subprocess.
type logLineMsg struct{ line string }

// installDoneMsg signals that an install subprocess finished.
type installDoneMsg struct{ exitCode int }

func (m *SetupModel) Init() tea.Cmd {
	m.subState = setupStateCheck
	return nil
}

func (m *SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		return m.handleKey(msg)

	case logLineMsg:
		m.logLines = append(m.logLines, msg.line)

	case installDoneMsg:
		if msg.exitCode == 0 {
			// Re-check after install
			m.brewMissing = !m.client.IsBrewInstalled()
			m.limaMissing = !m.client.IsLimactlInstalled()
			if !m.limaMissing {
				m.installDone = true
				m.subState = setupStateConfigCPUs
			} else {
				m.installError = "Installation failed — limactl still not found"
				m.subState = setupStateConfirm
			}
		} else {
			m.installError = fmt.Sprintf("Installation failed (exit %d)", msg.exitCode)
			m.subState = setupStateConfirm
		}
	}
	return m, nil
}

func (m *SetupModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.subState {
	case setupStateCheck, setupStateConfirm:
		switch msg.String() {
		case "y", "Y", "enter":
			m.subState = setupStateInstalling
			m.logLines = nil
			m.installError = ""
			return m, m.runInstall()
		case "n", "N", "q", "esc":
			return m, tea.Quit
		}

	case setupStateInstalling:
		// No key handling during install

	case setupStateConfigCPUs:
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
			m.subState = setupStateConfigRAM
		case "esc":
			m.subState = setupStateConfigCPUs
		}

	case setupStateConfigRAM:
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
			m.subState = setupStateConfigDisk
		case "esc":
			m.subState = setupStateConfigCPUs
		}

	case setupStateConfigDisk:
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
			// Save defaults and transition to main
			cpus := m.cpuOptions[m.cpuIdx]
			ram := m.ramOptions[m.ramIdx]
			disk := m.diskOptions[m.diskIdx]
			_ = m.client.WriteDefaultConfig(cpus, ram, disk)
			main := NewMainModel(m.client, m.width, m.height)
			return m, func() tea.Msg {
				return ChangeScreenMsg{State: StateMain, Screen: main}
			}
		case "esc":
			m.subState = setupStateConfigRAM
		}
	}
	return m, nil
}

// runInstall runs brew install (if needed) then lima install, streaming output.
func (m *SetupModel) runInstall() tea.Cmd {
	return func() tea.Msg {
		// Step 1: Install Homebrew if missing
		if m.brewMissing {
			cmd := exec.Command("bash", "-c",
				`NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)
			if err := streamCmd(cmd); err != nil {
				return installDoneMsg{exitCode: 1}
			}
		}
		// Step 2: Install Lima via brew
		cmd := exec.Command("brew", "install", "lima")
		if err := streamCmd(cmd); err != nil {
			return installDoneMsg{exitCode: 1}
		}
		return installDoneMsg{exitCode: 0}
	}
}

// streamCmd runs a command and discards output (simplified for non-streaming context).
// In a real TUI, output would be sent via tea.Cmd channels.
func streamCmd(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *SetupModel) View() string {
	var b strings.Builder

	title := TitleStyle.Render("aisand — Host Setup Required")
	b.WriteString(title + "\n\n")

	switch m.subState {
	case setupStateCheck, setupStateConfirm:
		b.WriteString("The following dependencies are required:\n\n")
		if m.brewMissing {
			b.WriteString(ErrorStyle.Render("  ✗ Homebrew (brew) — not found") + "\n")
		} else {
			b.WriteString(RunningBadgeStyle.Render("  ✓ Homebrew (brew) — found") + "\n")
		}
		if m.limaMissing {
			b.WriteString(ErrorStyle.Render("  ✗ Lima (limactl) — not found") + "\n")
		} else {
			b.WriteString(RunningBadgeStyle.Render("  ✓ Lima (limactl) — found") + "\n")
		}
		b.WriteString("\n")
		if m.installError != "" {
			b.WriteString(ErrorStyle.Render(m.installError) + "\n\n")
		}
		b.WriteString("Press " + KeyHintStyle.Render("y") + " to install, " + KeyHintStyle.Render("n") + " to quit.\n")

	case setupStateInstalling:
		b.WriteString("Installing dependencies...\n\n")
		// Show last 20 log lines
		start := 0
		if len(m.logLines) > 20 {
			start = len(m.logLines) - 20
		}
		for _, line := range m.logLines[start:] {
			b.WriteString(LogLineStyle.Render(line) + "\n")
		}
		if m.installDone {
			b.WriteString("\n" + RunningBadgeStyle.Render("✓ Installation complete!") + "\n")
		}

	case setupStateConfigCPUs:
		b.WriteString(RunningBadgeStyle.Render("✓ Lima installed successfully!") + "\n\n")
		b.WriteString("Configure default VM resources:\n\n")
		b.WriteString(TitleStyle.Render("Default CPUs:") + "\n")
		for i, v := range m.cpuOptions {
			if i == m.cpuIdx {
				b.WriteString(SelectedRowStyle.Render(fmt.Sprintf("  ▶ %d CPUs", v)) + "\n")
			} else {
				b.WriteString(NormalRowStyle.Render(fmt.Sprintf("    %d CPUs", v)) + "\n")
			}
		}
		b.WriteString("\n" + StatusBarStyle.Render("↑↓ select  enter confirm") + "\n")

	case setupStateConfigRAM:
		b.WriteString("Configure default VM resources:\n\n")
		b.WriteString(TitleStyle.Render("Default RAM:") + "\n")
		for i, v := range m.ramOptions {
			if i == m.ramIdx {
				b.WriteString(SelectedRowStyle.Render(fmt.Sprintf("  ▶ %d GB", v)) + "\n")
			} else {
				b.WriteString(NormalRowStyle.Render(fmt.Sprintf("    %d GB", v)) + "\n")
			}
		}
		b.WriteString("\n" + StatusBarStyle.Render("↑↓ select  enter confirm  esc back") + "\n")

	case setupStateConfigDisk:
		b.WriteString("Configure default VM resources:\n\n")
		b.WriteString(TitleStyle.Render("Default Disk:") + "\n")
		for i, v := range m.diskOptions {
			if i == m.diskIdx {
				b.WriteString(SelectedRowStyle.Render(fmt.Sprintf("  ▶ %d GB", v)) + "\n")
			} else {
				b.WriteString(NormalRowStyle.Render(fmt.Sprintf("    %d GB", v)) + "\n")
			}
		}
		b.WriteString("\n" + StatusBarStyle.Render("↑↓ select  enter confirm  esc back") + "\n")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
