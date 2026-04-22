package ui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
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
	setupStateAlreadyInstalled
	setupStateDone
)

// SetupModel is the onboarding screen shown when brew or limactl is not installed.
type SetupModel struct {
	client       *lima.Client
	subState     setupSubState
	brewMissing  bool
	limaMissing  bool
	installDone  bool
	installError string
	spinner      spinner.Model

	// Resource config selections
	cpuOptions []int
	ramOptions []int
	diskOptions []int
	cpuIdx     int
	ramIdx     int
	diskIdx    int

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

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	return &SetupModel{
		client:      client,
		brewMissing: !client.IsBrewInstalled(),
		limaMissing: !client.IsLimactlInstalled(),
		spinner:     s,
		cpuOptions:  cpuOptions,
		ramOptions:  []int{1, 2, 4, 8},
		diskOptions: []int{20, 40, 60, 80, 100},
		cpuIdx:      len(cpuOptions) - 1,
		ramIdx:      2,
		diskIdx:     2,
	}
}

// installDoneMsg signals that an install subprocess finished.
type installDoneMsg struct{ exitCode int }

func (m *SetupModel) Init() tea.Cmd {
	// If everything is already installed, show the "already installed" screen
	if !m.brewMissing && !m.limaMissing {
		m.subState = setupStateAlreadyInstalled
		return nil
	}
	m.subState = setupStateCheck
	return nil
}

func (m *SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		if m.subState == setupStateInstalling {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		return m.handleKey(msg)

	case installDoneMsg:
		// Re-check after install
		m.brewMissing = !m.client.IsBrewInstalled()
		m.limaMissing = !m.client.IsLimactlInstalled()
		if msg.exitCode == 0 && !m.limaMissing {
			m.installDone = true
			m.subState = setupStateConfigCPUs
		} else {
			if msg.exitCode != 0 {
				m.installError = fmt.Sprintf("Installation failed (exit %d). Press y to retry.", msg.exitCode)
			} else {
				m.installError = "Lima still not found after install. Press y to retry."
			}
			m.subState = setupStateConfirm
		}
	}
	return m, nil
}

func (m *SetupModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.subState {
	case setupStateAlreadyInstalled:
		// Only allow going back to main
		switch msg.String() {
		case "q", "esc", "enter":
			main := NewMainModel(m.client, m.width, m.height)
			return m, func() tea.Msg {
				return ChangeScreenMsg{State: StateMain, Screen: main}
			}
		}

	case setupStateCheck, setupStateConfirm:
		switch msg.String() {
		case "y", "Y", "enter":
			m.subState = setupStateInstalling
			m.installError = ""
			return m, tea.Batch(m.spinner.Tick, m.runInstall())
		case "n", "N", "q", "esc":
			return m, tea.Quit
		}

	case setupStateInstalling:
		// Block all keys during install

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

// runInstall installs missing dependencies idempotently.
func (m *SetupModel) runInstall() tea.Cmd {
	brewMissing := m.brewMissing
	limaMissing := m.limaMissing
	return func() tea.Msg {
		// Step 1: Install Homebrew if missing
		if brewMissing {
			cmd := exec.Command("bash", "-c",
				`NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)
			cmd.Stdout = nil
			cmd.Stderr = nil
			if err := cmd.Run(); err != nil {
				return installDoneMsg{exitCode: 1}
			}
		}
		// Step 2: Install Lima via brew if missing
		if limaMissing {
			cmd := exec.Command("brew", "install", "lima")
			cmd.Stdout = nil
			cmd.Stderr = nil
			if err := cmd.Run(); err != nil {
				return installDoneMsg{exitCode: 1}
			}
		}
		return installDoneMsg{exitCode: 0}
	}
}

func (m *SetupModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("aisand — Host Setup") + "\n\n")

	switch m.subState {
	case setupStateAlreadyInstalled:
		b.WriteString(RunningBadgeStyle.Render("  ✓ Homebrew (brew) — installed") + "\n")
		b.WriteString(RunningBadgeStyle.Render("  ✓ Lima (limactl) — installed") + "\n\n")
		b.WriteString(RunningBadgeStyle.Render("All dependencies are installed.") + "\n\n")
		b.WriteString(MutedStyle.Render("Press ") + KeyHintStyle.Render("q") + MutedStyle.Render(" or ") + KeyHintStyle.Render("esc") + MutedStyle.Render(" to go back.") + "\n")

	case setupStateCheck, setupStateConfirm:
		b.WriteString("The following dependencies are required to use aisand:\n\n")
		if m.brewMissing {
			b.WriteString(ErrorStyle.Render("  ✗ Homebrew (brew)  — not found") + "\n")
		} else {
			b.WriteString(RunningBadgeStyle.Render("  ✓ Homebrew (brew)  — installed") + "\n")
		}
		if m.limaMissing {
			b.WriteString(ErrorStyle.Render("  ✗ Lima (limactl)   — not found") + "\n")
		} else {
			b.WriteString(RunningBadgeStyle.Render("  ✓ Lima (limactl)   — installed") + "\n")
		}
		b.WriteString("\n")
		if m.installError != "" {
			b.WriteString(ErrorStyle.Render(m.installError) + "\n\n")
		}
		b.WriteString("Press " + KeyHintStyle.Render("y") + " to install missing dependencies, " + KeyHintStyle.Render("n") + " to quit.\n")

	case setupStateInstalling:
		b.WriteString(m.spinner.View() + " Installing dependencies — this may take a few minutes...\n\n")
		if m.brewMissing {
			b.WriteString(MutedStyle.Render("  • Installing Homebrew") + "\n")
		}
		if m.limaMissing {
			b.WriteString(MutedStyle.Render("  • Installing Lima (limactl)") + "\n")
		}
		b.WriteString("\n" + MutedStyle.Render("Please wait. Do not close this window.") + "\n")

	case setupStateConfigCPUs:
		b.WriteString(RunningBadgeStyle.Render("✓ Dependencies installed successfully!") + "\n\n")
		b.WriteString("Now configure default VM resources:\n\n")
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
