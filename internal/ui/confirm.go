package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmModel is a simple confirmation dialog.
type ConfirmModel struct {
	message   string
	onConfirm func() tea.Msg
	onCancel  func() tea.Msg
	width     int
	height    int
}

// NewConfirmModel creates a new ConfirmModel.
func NewConfirmModel(message string, onConfirm, onCancel func() tea.Msg) *ConfirmModel {
	return &ConfirmModel{
		message:   message,
		onConfirm: onConfirm,
		onCancel:  onCancel,
	}
}

func (m *ConfirmModel) Init() tea.Cmd { return nil }

func (m *ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			if m.onConfirm != nil {
				return m, func() tea.Msg { return m.onConfirm() }
			}
		case "n", "N", "esc":
			if m.onCancel != nil {
				return m, func() tea.Msg { return m.onCancel() }
			}
		}
	}
	return m, nil
}

func (m *ConfirmModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Confirm") + "\n\n")
	b.WriteString(NormalRowStyle.Render(m.message) + "\n\n")
	b.WriteString(
		KeyHintStyle.Render("[Y]") + " Yes   " +
			KeyHintStyle.Render("[N]") + " No\n",
	)

	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 3).
		Render(b.String())

	return lipgloss.NewStyle().Padding(2, 4).Render(dialog)
}
