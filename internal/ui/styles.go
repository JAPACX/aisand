package ui

import "github.com/charmbracelet/lipgloss"

// Color palette
const (
	ColorPrimary   = lipgloss.Color("#7C3AED") // purple — selected items
	ColorRunning   = lipgloss.Color("#10B981") // green — Running status
	ColorStopped   = lipgloss.Color("#6B7280") // gray — Stopped status
	ColorError     = lipgloss.Color("#EF4444") // red — errors
	ColorBorder    = lipgloss.Color("#374151") // dark gray — borders
	ColorStatusBar = lipgloss.Color("#1F2937") // very dark — status bar bg
	ColorText      = lipgloss.Color("#F9FAFB") // near white — primary text
	ColorMuted     = lipgloss.Color("#9CA3AF") // muted gray — secondary text
	ColorHighlight = lipgloss.Color("#4C1D95") // dark purple — selected row bg
	ColorTitle     = lipgloss.Color("#A78BFA") // light purple — titles
)

// Layout constants
const (
	LeftPanelRatio  = 0.35 // 35% of terminal width
	MinTermWidth    = 80
	MinTermHeight   = 24
	StatusBarHeight = 1
)

// Styles
var (
	// Panel border style
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	// Selected panel border (highlighted)
	ActivePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	// Selected row in a list
	SelectedRowStyle = lipgloss.NewStyle().
				Background(ColorHighlight).
				Foreground(ColorText).
				Bold(true)

	// Normal row in a list
	NormalRowStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// Title / heading
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorTitle).
			Bold(true).
			Padding(0, 1)

	// Status badge — Running
	RunningBadgeStyle = lipgloss.NewStyle().
				Foreground(ColorRunning).
				Bold(true)

	// Status badge — Stopped
	StoppedBadgeStyle = lipgloss.NewStyle().
				Foreground(ColorStopped)

	// Status badge — other/unknown
	UnknownBadgeStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	// Log line (normal output)
	LogLineStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Error text
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// Status bar at the bottom
	StatusBarStyle = lipgloss.NewStyle().
			Background(ColorStatusBar).
			Foreground(ColorMuted).
			Padding(0, 1)

	// Key hint in status bar
	KeyHintStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// Muted / disabled item
	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Detail label (e.g. "CPUs:", "RAM:")
	DetailLabelStyle = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Width(10)

	// Detail value
	DetailValueStyle = lipgloss.NewStyle().
				Foreground(ColorText)
)

// StatusBadge returns a styled status string for a VM status.
func StatusBadge(status string) string {
	switch status {
	case "Running":
		return RunningBadgeStyle.Render("● " + status)
	case "Stopped":
		return StoppedBadgeStyle.Render("○ " + status)
	default:
		return UnknownBadgeStyle.Render("? " + status)
	}
}
