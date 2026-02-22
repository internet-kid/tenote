package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/internet-kid/tenote/internal/ui/app"
	"github.com/internet-kid/tenote/internal/ui/menu"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	p := tea.NewProgram(newRoot(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return nil
}

// ── root model ────────────────────────────────────────────────────────────────

type rootState int

const (
	atMenu rootState = iota
	atApp
)

// root is the top-level tea.Model.
// It starts on the animated main menu and switches to the note app
// when the user selects "Open Notes".
type root struct {
	state  rootState
	menu   menu.Model
	app    app.Model
	width  int
	height int
}

func newRoot() root {
	return root{menu: menu.New()}
}

func (r root) Init() tea.Cmd {
	return r.menu.Init()
}

func (r root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Always track window size so we can forward it on screen transitions.
	if sz, ok := msg.(tea.WindowSizeMsg); ok {
		r.width = sz.Width
		r.height = sz.Height
	}

	// Switch from menu → app when the user picks "Open Notes".
	if _, ok := msg.(menu.OpenNotesMsg); ok {
		appModel, err := app.NewModel()
		if err != nil {
			// Stay on the menu if the app fails to initialise.
			return r, nil
		}
		r.state = atApp
		r.app = appModel
		// Seed the app with the current terminal dimensions.
		next, cmd := r.app.Update(tea.WindowSizeMsg{Width: r.width, Height: r.height})
		r.app = next.(app.Model)
		return r, cmd
	}

	if r.state == atApp {
		next, cmd := r.app.Update(msg)
		r.app = next.(app.Model)
		return r, cmd
	}

	next, cmd := r.menu.Update(msg)
	r.menu = next.(menu.Model)
	return r, cmd
}

func (r root) View() string {
	if r.state == atApp {
		return r.app.View()
	}
	return r.menu.View()
}
