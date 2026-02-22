package menu

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OpenNotesMsg is sent to the parent model when the user picks "Open Notes".
type OpenNotesMsg struct{}

type tickMsg time.Time

type viewState int

const (
	viewAnim     viewState = iota // logo is being drawn line by line
	viewMenu                      // main menu is visible
	viewInfo                      // info / about screen
	viewSettings                  // settings screen
)

// logo is the ASCII art for "tenote" in ANSI Shadow style.
var logo = []string{
	`████████╗███████╗███╗  ██╗ ██████╗ ████████╗███████╗`,
	`╚══██╔══╝██╔════╝████╗ ██║██╔═══██╗╚══██╔══╝██╔════╝`,
	`   ██║   █████╗  ██╔██╗██║██║   ██║   ██║   █████╗  `,
	`    ██║   ██╔══╝  ██║╚██╗██║██║   ██║  ██║   ██╔══╝  `,
	`   ██║   ███████╗██║ ╚████║╚██████╔╝  ██║   ███████╗`,
	`   ╚═╝   ╚══════╝╚═╝  ╚═══╝ ╚═════╝   ╚═╝   ╚══════╝`,
}


type menuItem struct {
	label string
	id    int
}

const (
	idNotes = iota
	idSettings
	idInfo
	idQuit
)

var menuItems = []menuItem{
	{"Open Notes", idNotes},
	{"Settings", idSettings},
	{"Information", idInfo},
	{"Exit", idQuit},
}

// Model is the main-menu tea.Model.
type Model struct {
	width, height int
	view          viewState
	linesShown int // how many logo lines are currently visible
	cursor     int // selected menu item
}

// New returns a fresh Model ready to animate.
func New() Model { return Model{} }

func (m Model) Init() tea.Cmd {
	return doTick(80 * time.Millisecond)
}

func doTick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tickMsg:
		return m.onTick()
	case tea.KeyMsg:
		return m.onKey(msg)
	}
	return m, nil
}

func (m Model) onTick() (Model, tea.Cmd) {
	switch m.view {
	case viewAnim:
		m.linesShown++
		if m.linesShown >= len(logo) {
			m.view = viewMenu
			return m, doTick(120 * time.Millisecond)
		}
		return m, doTick(80 * time.Millisecond)

	}
	return m, nil
}

func (m Model) onKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	// Universal quit.
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.view {
	case viewAnim:
		// Any key skips the animation.
		m.linesShown = len(logo)
		m.view = viewMenu
		return m, doTick(120 * time.Millisecond)

	case viewMenu:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(menuItems)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m.pick()
		case "q":
			return m, tea.Quit
		}

	case viewInfo, viewSettings:
		if msg.String() == "esc" || msg.String() == "q" {
			m.view = viewMenu
		}
	}

	return m, nil
}

func (m Model) pick() (Model, tea.Cmd) {
	switch menuItems[m.cursor].id {
	case idNotes:
		return m, func() tea.Msg { return OpenNotesMsg{} }
	case idSettings:
		m.view = viewSettings
	case idInfo:
		m.view = viewInfo
	case idQuit:
		return m, tea.Quit
	}
	return m, nil
}

// ── styles ────────────────────────────────────────────────────────────────────

var (
	subtitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#25b067")).Bold(true)
	activeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	hintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	panelStyle    = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 3)
	headStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#25b067"))
	boldStyle = lipgloss.NewStyle().Bold(true)
)

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	switch m.view {
	case viewInfo:
		return m.viewInfo()
	case viewSettings:
		return m.viewSettings()
	default:
		return m.viewMain()
	}
}

func (m Model) renderLogo() string {
	ls := lipgloss.NewStyle().Foreground(lipgloss.Color("#25b067")).Bold(true)
	var b strings.Builder
	for i, line := range logo {
		if i < m.linesShown {
			b.WriteString(ls.Render(line))
		}
		if i < len(logo)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m Model) viewMain() string {
	parts := []string{m.renderLogo()}

	if m.view == viewMenu {
		parts = append(parts,
			"",
			subtitleStyle.Render("your notes, your way"),
			"",
		)

		var mb strings.Builder
		for i, item := range menuItems {
			if i == m.cursor {
				mb.WriteString(cursorStyle.Render(" →  ") + activeStyle.Render(item.label))
			} else {
				mb.WriteString(dimStyle.Render("    "+item.label))
			}
			if i < len(menuItems)-1 {
				mb.WriteString("\n")
			}
		}

		parts = append(parts,
			mb.String(),
			"",
			hintStyle.Render("↑↓ / jk  navigate  •  enter  select  •  q  quit"),
		)
	}

	content := lipgloss.JoinVertical(lipgloss.Center, parts...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m Model) viewInfo() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		headStyle.Render("Information"),
		"",
		"A minimal TUI note-taking application.",
		"",
		boldStyle.Render("Sections"),
		"  Notes    — regular notes",
		"  Trash    — deleted notes",
		"",
		boldStyle.Render("Shortcuts"),
		"  ↑↓ / jk   navigate list",
		"  n          new note (Notes only)",
		"  e          edit note (Notes only)",
		"  d          move to trash / delete forever",
		"  r          restore from trash",
		"  Ctrl+S     save",
		"  ?          toggle help",
		"  Tab        switch focus",
		"",
		hintStyle.Render("q / esc — back to menu"),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		panelStyle.Render(body))
}

func (m Model) viewSettings() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		headStyle.Render("Settings"),
		"",
		dimStyle.Render("Nothing to configure yet."),
		dimStyle.Render("This section will grow in future versions."),
		"",
		hintStyle.Render("q / esc — back to menu"),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		panelStyle.Render(body))
}
