package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/japacx/aisand/internal/lima"
)

// AppState represents the current screen state of the application.
type AppState int

const (
	StateSetup AppState = iota
	StateMain
	StateActionMenu
	StateGlobalMenu
	StateCreateVM
	StateLogView
	StateConfirm
	StateMountsList
	StateUnmountPicker
	StateToolPicker
)

// App is the root Bubbletea model. It holds the current screen and delegates
// Update/View to it. Screen transitions happen by replacing currentScreen.
type App struct {
	state         AppState
	currentScreen tea.Model
	width         int
	height        int
	limaClient    *lima.Client
}

// NewApp creates a new App model.
func NewApp(client *lima.Client) *App {
	return &App{
		limaClient: client,
	}
}

// Init checks for brew + limactl and sets the initial screen.
func (a *App) Init() tea.Cmd {
	if !a.limaClient.IsBrewInstalled() || !a.limaClient.IsLimactlInstalled() {
		a.state = StateSetup
		setup := NewSetupModel(a.limaClient)
		a.currentScreen = setup
		return setup.Init()
	}
	a.state = StateMain
	main := NewMainModel(a.limaClient, a.width, a.height)
	a.currentScreen = main
	return main.Init()
}

// Update delegates to the current screen and handles global messages.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Propagate size to current screen
		if a.currentScreen != nil {
			newScreen, cmd := a.currentScreen.Update(msg)
			a.currentScreen = newScreen
			return a, cmd
		}
		return a, nil

	case tea.KeyMsg:
		// Global ctrl+c handler
		if msg.Type == tea.KeyCtrlC {
			return a, tea.Quit
		}

	case ChangeScreenMsg:
		return a.handleScreenChange(msg)

	case resumeLogViewMsg:
		// Restore the log view as the current screen (user chose not to cancel)
		a.state = StateLogView
		a.currentScreen = msg.logView
		return a, nil
	}

	// Delegate to current screen
	if a.currentScreen != nil {
		newScreen, cmd := a.currentScreen.Update(msg)
		a.currentScreen = newScreen
		return a, cmd
	}
	return a, nil
}

// View delegates rendering to the current screen.
func (a *App) View() string {
	if a.currentScreen == nil {
		return "Loading..."
	}
	return a.currentScreen.View()
}

// ChangeScreenMsg is sent by screens to request a state transition.
type ChangeScreenMsg struct {
	State  AppState
	Screen tea.Model
}

// handleScreenChange transitions to a new screen.
func (a *App) handleScreenChange(msg ChangeScreenMsg) (tea.Model, tea.Cmd) {
	a.state = msg.State
	a.currentScreen = msg.Screen
	if msg.Screen != nil {
		return a, msg.Screen.Init()
	}
	return a, nil
}
