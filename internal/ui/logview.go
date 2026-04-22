package ui

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/japacx/aisand/internal/lima"
)

// logLineMsgLV carries a new output line from the subprocess.
type logLineMsgLV struct{ line string }

// processDoneMsg signals that the subprocess has exited.
type processDoneMsg struct{ exitCode int }

// LogViewModel streams subprocess output in a scrollable viewport.
type LogViewModel struct {
	client   *lima.Client
	cmd      *exec.Cmd
	title    string
	viewport viewport.Model
	lines    []string
	done     bool
	exitCode int
	onDone   func(int) tea.Msg
	width    int
	height   int
	scanner  *bufio.Scanner
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
		onDone:   onDone,
		width:    width,
		height:   height,
	}
}

// Init starts the subprocess and begins streaming output.
func (m *LogViewModel) Init() tea.Cmd {
	return func() tea.Msg {
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
		// Store scanner for subsequent reads in Update
		m.scanner = scanner
		return m.readNextLine()
	}
}

// readNextLine reads the next line from the scanner, returning either a
// logLineMsgLV (if a line was read) or a processDoneMsg (when the stream ends).
func (m *LogViewModel) readNextLine() tea.Msg {
	if m.scanner != nil && m.scanner.Scan() {
		return logLineMsgLV{line: m.scanner.Text()}
	}
	// Scanner exhausted — process has finished writing. Collect exit code.
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

	case logLineMsgLV:
		if msg.line != "" {
			m.lines = append(m.lines, msg.line)
			m.viewport.SetContent(m.renderLines())
			m.viewport.GotoBottom()
		}
		// Schedule next line read — this is what drives real-time streaming.
		return m, func() tea.Msg { return m.readNextLine() }

	case processDoneMsg:
		m.done = true
		m.exitCode = msg.exitCode
		return m, nil

	case tea.KeyMsg:
		if m.done {
			// Any key press after completion calls onDone
			if m.onDone != nil {
				onDone := m.onDone
				exitCode := m.exitCode
				return m, func() tea.Msg {
					return onDone(exitCode)
				}
			}
			return m, nil
		}
		// Allow viewport scrolling while running
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
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
	b.WriteString(m.viewport.View() + "\n")

	if m.done {
		if m.exitCode == 0 {
			b.WriteString("\n" + RunningBadgeStyle.Render("✓ Done. Press any key to continue.") + "\n")
		} else {
			b.WriteString("\n" + ErrorStyle.Render(fmt.Sprintf("✗ Failed (exit %d). Press any key to continue.", m.exitCode)) + "\n")
		}
	} else {
		b.WriteString("\n" + MutedStyle.Render("Running... (↑↓ to scroll)") + "\n")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
