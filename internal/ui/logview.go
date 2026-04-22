package ui

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

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

// elapsedTickMsg is sent every second to update the elapsed timer.
type elapsedTickMsg struct{}

// resumeLogViewMsg is sent when the user cancels the cancel-dialog and wants
// to return to the running log view.
type resumeLogViewMsg struct{ logView *LogViewModel }

// LogViewModel streams subprocess output in a scrollable viewport.
type LogViewModel struct {
	client     *lima.Client
	cmd        *exec.Cmd
	title      string
	hint       string // optional context message shown below the title
	vmName     string
	isCreating bool
	viewport   viewport.Model
	spinner    spinner.Model
	lines      []string
	done       bool
	exitCode   int
	elapsed    int
	startTime  time.Time
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

// NewLogViewModelWithHint creates a LogViewModel with an optional hint message.
func NewLogViewModelWithHint(client *lima.Client, cmd *exec.Cmd, title, hint string, width, height int, onDone func(int) tea.Msg) *LogViewModel {
	vp := viewport.New(width, height-6)
	vp.SetContent("")
	return &LogViewModel{
		client:   client,
		cmd:      cmd,
		title:    title,
		hint:     hint,
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
		hint:       "This operation may take a moment, please wait.",
		vmName:     vmName,
		isCreating: true,
		viewport:   vp,
		spinner:    newSpinner(),
		onDone:     onDone,
		width:      width,
		height:     height,
	}
}

func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return elapsedTickMsg{}
	})
}

// Init starts the subprocess, begins streaming output, and starts the spinner + timer.
func (m *LogViewModel) Init() tea.Cmd {
	m.startTime = time.Now()
	return tea.Batch(
		m.spinner.Tick,
		tickEverySecond(),
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

	case elapsedTickMsg:
		if !m.done {
			m.elapsed = int(time.Since(m.startTime).Seconds())
			return m, tickEverySecond()
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
		m.elapsed = int(time.Since(m.startTime).Seconds())
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

// fmtElapsed formats seconds into a human-readable string.
func fmtElapsed(secs int) string {
	if secs < 60 {
		return fmt.Sprintf("%ds elapsed", secs)
	}
	return fmt.Sprintf("%dm%ds elapsed", secs/60, secs%60)
}

// View renders the log view.
func (m *LogViewModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(m.title) + "\n")

	if m.hint != "" {
		b.WriteString(MutedStyle.Render("ℹ  "+m.hint) + "\n")
	}

	b.WriteString("\n")

	if !m.done && len(m.lines) == 0 {
		// hint already shown above — nothing extra needed while waiting
	} else {
		b.WriteString(m.viewport.View() + "\n")
	}

	if m.done {
		if m.exitCode == 0 {
			b.WriteString("\n" + RunningBadgeStyle.Render(
				fmt.Sprintf("✓ Done in %s. Press any key to continue.", fmtElapsed(m.elapsed)),
			) + "\n")
		} else {
			b.WriteString("\n" + ErrorStyle.Render(
				fmt.Sprintf("✗ Failed (exit %d) after %s. Press any key to continue.", m.exitCode, fmtElapsed(m.elapsed)),
			) + "\n")
		}
	} else {
		hint := "Running..."
		if m.isCreating {
			hint = "Running... (esc to cancel)"
		}
		elapsedStr := MutedStyle.Render(fmtElapsed(m.elapsed))
		b.WriteString("\n" + MutedStyle.Render(m.spinner.View()+" "+hint) + "  " + elapsedStr + "\n")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
