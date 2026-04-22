package ui

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/japacx/aisand/internal/lima"
)

// logLineMsgLV carries a new output line from the subprocess.
type logLineMsgLV struct{ line string }

// processDoneMsg signals that the subprocess has exited.
type processDoneMsg struct{ exitCode int }

// resumeLogViewMsg is sent when the user cancels the cancel-dialog and wants
// to return to the running log view.
type resumeLogViewMsg struct{ logView *LogViewModel }

// LogViewModel streams subprocess output in a scrollable viewport.
type LogViewModel struct {
	client     *lima.Client
	cmd        *exec.Cmd
	title      string
	vmName     string
	isCreating bool
	viewport   viewport.Model
	spinner    spinner.Model
	lines      []string
	done       bool
	exitCode   int
	onDone     func(int) tea.Msg
	width      int
	height     int
	scanner    *bufio.Scanner
}

func newSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)
	return s
}

// NewLogViewModel creates a new LogViewModel.
func NewLogViewModel(client *lima.Client, cmd *exec.Cmd, title string, width, height int, onDone func(int) tea.Msg) *LogViewModel {
	vp := viewport.New(width, height-6)
	vp.SetContent("")
	return &LogViewModel{
		client:   client,
		cmd:      cmd,
		title:    title,
		viewport: vp,
		spinner:  newSpinner(),
		onDone:   onDone,
		width:    width,
		height:   height,
	}
}

// NewCreationLogViewModel creates a LogViewModel for VM creation with cancel support.
func NewCreationLogViewModel(client *lima.Client, cmd *exec.Cmd, vmName string, width, height int, onDone func(int) tea.Msg) *LogViewModel {
	vp := viewport.New(width, height-6)
	vp.SetContent("")
	return &LogViewModel{
		client:     client,
		cmd:        cmd,
		title:      fmt.Sprintf("Creating VM %q...", vmName),
		vmName:     vmName,
		isCreating: true,
		viewport:   vp,
		spinner:    newSpinner(),
		onDone:     onDone,
		width:      width,
		height:     height,
	}
}

// Init starts the subprocess, begins streaming output, and starts the spinner.
func (m *LogViewModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			stdout, err := m.cmd.StdoutPipe()
			if err != nil {
				return processDoneMsg{exitCode: 1}
			}
			stderr, err := m.cmd.StderrPipe()
			if err != nil {
				return processDoneMsg{exitCode: 1}
			}
			if err := m.cmd.Start(); err != nil {
				return processDoneMsg{exitCode: 1}
			}
			combined := io.MultiReader(stdout, stderr)
			scanner := bufio.NewScanner(combined)
			m.scanner = scanner
			return m.readNextLine()
		},
	)
}

// readNextLine reads the next line from the scanner.
func (m *LogViewModel) readNextLine() tea.Msg {
	if m.scanner != nil && m.scanner.Scan() {
		return logLineMsgLV{line: m.scanner.Text()}
	}
	exitCode := 0
	if m.cmd.ProcessState != nil {
		exitCode = m.cmd.ProcessState.ExitCode()
	} else if err := m.cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}
	return processDoneMsg{exitCode: exitCode}
}

// Update handles incoming messages and drives the streaming loop.
func (m *LogViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6
		return m, nil

	case spinner.TickMsg:
		if !m.done {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case logLineMsgLV:
		if msg.line != "" {
			m.lines = append(m.lines, msg.line)
			m.viewport.SetContent(m.renderLines())
			m.viewport.GotoBottom()
		}
		return m, func() tea.Msg { return m.readNextLine() }

	case processDoneMsg:
		m.done = true
		m.exitCode = msg.exitCode
		return m, nil

	case resumeLogViewMsg:
		return m, func() tea.Msg {
			return ChangeScreenMsg{State: StateLogView, Screen: msg.logView}
		}

	case tea.KeyMsg:
		if m.done {
			if m.onDone != nil {
				onDone := m.onDone
				exitCode := m.exitCode
				return m, func() tea.Msg { return onDone(exitCode) }
			}
			return m, nil
		}

		if msg.Type == tea.KeyEsc && m.isCreating {
			self := m
			vmName := m.vmName
			client := m.client
			width := m.width
			height := m.height

			confirm := NewConfirmModel(
				fmt.Sprintf("Cancel creation of VM %q?\n\nIf you confirm, the VM will be deleted.", vmName),
				func() tea.Msg {
					if self.cmd.Process != nil {
						_ = self.cmd.Process.Kill()
					}
					_ = client.DeleteVM(vmName)
					main := NewMainModel(client, width, height)
					return ChangeScreenMsg{State: StateMain, Screen: main}
				},
				func() tea.Msg {
					return resumeLogViewMsg{logView: self}
				},
			)
			return m, func() tea.Msg {
				return ChangeScreenMsg{State: StateConfirm, Screen: confirm}
			}
		}

		return m, nil
	}

	return m, nil
}

// renderLines renders all collected log lines into a single string for the viewport.
func (m *LogViewModel) renderLines() string {
	var b strings.Builder
	for _, line := range m.lines {
		b.WriteString(LogLineStyle.Render(line) + "\n")
	}
	return b.String()
}

// View renders the log view.
func (m *LogViewModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(m.title) + "\n\n")

	if !m.done && len(m.lines) == 0 {
		// No output yet — show spinner while waiting
		b.WriteString(m.spinner.View() + " " + MutedStyle.Render("Waiting for output...") + "\n")
	} else {
		b.WriteString(m.viewport.View() + "\n")
	}

	if m.done {
		if m.exitCode == 0 {
			b.WriteString("\n" + RunningBadgeStyle.Render("✓ Done. Press any key to continue.") + "\n")
		} else {
			b.WriteString("\n" + ErrorStyle.Render(fmt.Sprintf("✗ Failed (exit %d). Press any key to continue.", m.exitCode)) + "\n")
		}
	} else {
		hint := "Running..."
		if m.isCreating {
			hint = "Running... (esc to cancel)"
		}
		b.WriteString("\n" + MutedStyle.Render(m.spinner.View()+" "+hint) + "\n")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
