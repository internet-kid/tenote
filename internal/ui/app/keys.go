package app

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	// global
	Quit key.Binding
	Help key.Binding
	Tab  key.Binding

	// browse
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	SectionUp key.Binding
	SectionDn key.Binding
	New       key.Binding
	Edit      key.Binding
	Trash     key.Binding
	Delete    key.Binding
	Restore   key.Binding

	// edit mode
	Save   key.Binding
	Cancel key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "focus"),
		),

		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),

		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "sidebar"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "preview"),
		),

		SectionUp: key.NewBinding(
			key.WithKeys("K"),
			key.WithHelp("K", "prev section"),
		),
		SectionDn: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "next section"),
		),

		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new note"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Trash: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "to trash"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete forever"),
		),
		Restore: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restore"),
		),

		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help,
		k.Tab,
		k.New,
		k.Edit,
		k.Trash,
		k.Quit,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.SectionUp, k.SectionDn},
		{k.New, k.Edit},
		{k.Trash, k.Restore},
		{k.Tab, k.Help},
		{k.Quit},
	}
}

func (k KeyMap) EditShortHelp() []key.Binding {
	return []key.Binding{
		k.Save,
		k.Cancel,
		k.Quit,
	}
}

func (k KeyMap) TrashShortHelp() []key.Binding {
	return []key.Binding{
		k.Delete,
		k.Restore,
		k.Quit,
	}
}
