package menu

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/internet-kid/tenote/internal/config"
)

// OpenNotesMsg is sent to the parent model when the user picks "Open Notes".
type OpenNotesMsg struct{}

type tickMsg time.Time

type viewState int

const (
	viewAnim       viewState = iota // logo is being drawn line by line
	viewMenu                        // main menu is visible
	viewInfo                        // info / about screen
	viewSettings                    // settings screen
	viewFilePicker                  // folder picker inside settings
	viewMkdir                       // new-folder dialog (opened from file picker)
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
	linesShown    int // how many logo lines are currently visible
	cursor        int // selected menu item

	input    textinput.Model
	inputErr string

	fp filepicker.Model

	mkdirInput textinput.Model
	mkdirErr   string
}

// New returns a fresh Model ready to animate.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "e.g. ~/.local/share/tenote"
	ti.CharLimit = 512
	ti.Width = 52

	cfg, _ := config.LoadConfig()
	ti.SetValue(cfg.StorageDir)

	mi := textinput.New()
	mi.Placeholder = "folder name"
	mi.CharLimit = 255
	mi.Width = 52

	return Model{input: ti, mkdirInput: mi}
}

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

	// Forward non-key messages to the filepicker when active.
	if m.view == viewFilePicker {
		var cmd tea.Cmd
		m.fp, cmd = m.fp.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) onTick() (Model, tea.Cmd) {
	if m.view == viewAnim {
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

	case viewInfo:
		if msg.String() == "esc" || msg.String() == "q" {
			m.view = viewMenu
		}

	case viewSettings:
		switch msg.String() {
		case "esc":
			cfg, _ := config.LoadConfig()
			m.input.SetValue(cfg.StorageDir)
			m.inputErr = ""
			m.input.Blur()
			m.view = viewMenu
		case "enter":
			dir := expandTilde(strings.TrimSpace(m.input.Value()))
			if dir == "" {
				m.inputErr = "Path cannot be empty"
				return m, nil
			}
			if err := config.SaveConfig(config.AppConfig{StorageDir: dir}); err != nil {
				m.inputErr = err.Error()
				return m, nil
			}
			m.input.SetValue(dir)
			m.inputErr = ""
			m.input.Blur()
			m.view = viewMenu
		case "ctrl+f":
			m.input.Blur()
			m = m.initFilePicker()
			return m, m.fp.Init()
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

	case viewFilePicker:
		switch msg.String() {
		case "esc":
			m.view = viewSettings
			return m, m.input.Focus()
		case "ctrl+n":
			m.mkdirInput.SetValue("")
			m.mkdirErr = ""
			m.view = viewMkdir
			return m, m.mkdirInput.Focus()
		default:
			var cmd tea.Cmd
			m.fp, cmd = m.fp.Update(msg)

			if ok, path := m.fp.DidSelectFile(msg); ok {
				m.input.SetValue(path)
				m.inputErr = ""
				m.view = viewSettings
				return m, m.input.Focus()
			}
			return m, cmd
		}

	case viewMkdir:
		switch msg.String() {
		case "esc":
			m.mkdirInput.Blur()
			m.mkdirErr = ""
			m.view = viewFilePicker
		case "enter":
			name := strings.TrimSpace(m.mkdirInput.Value())
			if name == "" {
				m.mkdirErr = "Folder name cannot be empty"
				return m, nil
			}
			newDir := filepath.Join(m.fp.CurrentDirectory, name)
			if err := os.MkdirAll(newDir, 0o755); err != nil {
				m.mkdirErr = err.Error()
				return m, nil
			}
			// Select the new folder and return to settings.
			m.input.SetValue(newDir)
			m.inputErr = ""
			m.mkdirInput.Blur()
			m.mkdirErr = ""
			m.view = viewSettings
			return m, m.input.Focus()
		default:
			var cmd tea.Cmd
			m.mkdirInput, cmd = m.mkdirInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) pick() (Model, tea.Cmd) {
	switch menuItems[m.cursor].id {
	case idNotes:
		return m, func() tea.Msg { return OpenNotesMsg{} }
	case idSettings:
		cfg, _ := config.LoadConfig()
		m.input.SetValue(cfg.StorageDir)
		m.inputErr = ""
		m.view = viewSettings
		return m, m.input.Focus()
	case idInfo:
		m.view = viewInfo
	case idQuit:
		return m, tea.Quit
	}
	return m, nil
}

// initFilePicker sets up the filepicker, starting from the current input value.
func (m Model) initFilePicker() Model {
	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.ShowHidden = false

	startDir := expandTilde(strings.TrimSpace(m.input.Value()))
	if info, err := os.Stat(startDir); err != nil || !info.IsDir() {
		if home, err := os.UserHomeDir(); err == nil {
			startDir = home
		}
	}
	fp.CurrentDirectory = startDir

	pickerH := max(m.height-12, 5)
	fp.SetHeight(pickerH)

	m.fp = fp
	m.view = viewFilePicker
	return m
}

// expandTilde replaces a leading ~ with the user's home directory.
func expandTilde(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
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
	headStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#25b067"))
	boldStyle  = lipgloss.NewStyle().Bold(true)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
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
	case viewFilePicker:
		return m.viewFilePicker()
	case viewMkdir:
		return m.viewMkdir()
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
	rows := []string{
		headStyle.Render("Settings"),
		"",
		boldStyle.Render("Storage directory"),
		dimStyle.Render("Notes and trash are stored inside this folder"),
		"",
		m.input.View(),
		"",
	}

	if m.inputErr != "" {
		rows = append(rows, errorStyle.Render(m.inputErr), "")
	}

	rows = append(rows, hintStyle.Render("enter  save  •  ctrl+f  browse  •  esc  cancel"))

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		panelStyle.Render(body))
}

func (m Model) viewFilePicker() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		headStyle.Render("Browse"),
		dimStyle.Render("Navigate to a folder and press enter to select it"),
		"",
		m.fp.View(),
		"",
		hintStyle.Render("↑↓ / jk  navigate  •  enter  open/select  •  backspace  back  •  ctrl+n  new folder  •  esc  cancel"),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		panelStyle.Render(body))
}

func (m Model) viewMkdir() string {
	rows := []string{
		headStyle.Render("New Folder"),
		dimStyle.Render("Creating inside: " + m.fp.CurrentDirectory),
		"",
		m.mkdirInput.View(),
		"",
	}

	if m.mkdirErr != "" {
		rows = append(rows, errorStyle.Render(m.mkdirErr), "")
	}

	rows = append(rows, hintStyle.Render("enter  create  •  esc  cancel"))

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		panelStyle.Render(body))
}
