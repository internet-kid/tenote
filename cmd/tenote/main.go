package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"tenote/internal/ui/app"
)

func main() {
	m, err := app.NewModel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "init error:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "run error:", err)
		os.Exit(1)
	}
}
