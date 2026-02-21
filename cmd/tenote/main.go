package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"tenote/internal/ui/app"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	m, err := app.NewModel()
	if err != nil {
		return fmt.Errorf("init app model: %w", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run app: %w", err)
	}

	return nil
}
