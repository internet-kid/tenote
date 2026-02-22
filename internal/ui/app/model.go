package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"tenote/internal/config"
	"tenote/internal/storage/fs"
)

type focusArea int
type mode int

const (
	timeLayout = "2006-01-02 15:04"
	loadingMsg = "loading..."
)

const (
	focusSidebar focusArea = iota
	focusPreview
)

const (
	modeBrowse mode = iota
	modeEdit
)

type sectionItem struct {
	key   fs.Section
	title string
}

var sections = []sectionItem{
	{key: fs.SectionNotes, title: "Notes"},
	{key: fs.SectionTrash, title: "Trash"},
}

type noteItem struct {
	n fs.Note
}

func (i noteItem) Title() string       { return i.n.Title }
func (i noteItem) Description() string { return i.n.UpdatedAt.Format(timeLayout) }
func (i noteItem) FilterValue() string { return i.n.Title }

type Model struct {
	store *fs.Store

	width  int
	height int

	focus focusArea

	mode   mode
	editor textarea.Model

	dirty   bool
	editErr error

	sectionIdx int
	noteList   list.Model
	preview    viewport.Model

	notes      []fs.Note
	selected   *fs.Note
	previewErr error

	help     help.Model
	keys     KeyMap
	showHelp bool

	status string

	renderer *glamour.TermRenderer
}

func NewModel() (Model, error) {
	paths, err := config.ResolvePaths()
	if err != nil {
		return Model{}, err
	}
	store := fs.NewStore(paths)

	del := list.NewDefaultDelegate()
	del.Styles.SelectedTitle = del.Styles.SelectedTitle.Foreground(lipgloss.Color("#25b067")).BorderForeground(lipgloss.Color("#25b067"))
	del.Styles.SelectedDesc = del.Styles.SelectedDesc.Foreground(lipgloss.Color("#25b067")).BorderForeground(lipgloss.Color("#25b067"))
	l := list.New([]list.Item{}, del, 0, 0)
	l.Title = "Notes"
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	vp := viewport.New(0, 0)
	vp.SetContent("")

	ta := textarea.New()
	ta.Placeholder = "Write your note..."
	ta.ShowLineNumbers = true
	ta.Prompt = ""
	ta.CharLimit = 0
	h := help.New()
	h.ShowAll = false

	m := Model{
		store:      store,
		focus:      focusSidebar,
		sectionIdx: 0,
		noteList:   l,
		preview:    vp,
		mode:       modeBrowse,
		editor:     ta,
		help:       h,
		keys:       DefaultKeyMap(),
		showHelp:   false,
	}

	if err := m.reloadNotes(); err != nil {
		m.status = "load error: " + err.Error()
	}
	m.syncSelection()
	return m, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout()
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

		if m.mode == modeEdit {
			next, cmd := m.updateEditMode(msg)
			return next, cmd
		}

		next, cmd := m.updateBrowseMode(msg)
		return next, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return loadingMsg
	}

	sidebar := m.renderSidebar()
	preview := m.renderPreview()

	root := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, preview)
	status := m.renderStatus()
	help := m.renderHelp()

	return lipgloss.JoinVertical(lipgloss.Left, root, status, help)
}

// ---------- rendering ----------

var (
	border = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))

	titleStyle  = lipgloss.NewStyle().Bold(true)
	blurStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	focusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#25b067"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
)

func (m Model) updateEditMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.exitEditMode("Canceled")
		return m, nil

	case key.Matches(msg, m.keys.Save):
		if m.selected == nil {
			return m, nil
		}

		selectedID := m.selected.ID
		body := m.editor.Value()
		if err := m.store.WriteBody(m.selected.Path, body); err != nil {
			m.status = "save error: " + err.Error()
			m.editErr = err
			return m, nil
		}

		m.exitEditMode("Saved")
		m.refreshNotesAndReselect(selectedID)
		return m, nil
	}

	var cmd tea.Cmd
	before := m.editor.Value()
	m.editor, cmd = m.editor.Update(msg)
	if m.editor.Value() != before {
		m.dirty = true
	}

	return m, cmd
}

func (m Model) updateBrowseMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Help):
		if sections[m.sectionIdx].key == fs.SectionTrash {
			return m, nil
		}
		m.showHelp = !m.showHelp
		m.help.ShowAll = m.showHelp
		return m, nil

	case key.Matches(msg, m.keys.Tab):
		if m.focus == focusSidebar {
			m.focus = focusPreview
		} else {
			m.focus = focusSidebar
		}
		return m, nil

	case key.Matches(msg, m.keys.Left):
		m.focus = focusSidebar
		return m, nil

	case key.Matches(msg, m.keys.Right):
		m.focus = focusPreview
		return m, nil

	case key.Matches(msg, m.keys.SectionDn):
		m.sectionIdx = min(m.sectionIdx+1, len(sections)-1)
		m.refreshNotesAndSelection()
		return m, nil

	case key.Matches(msg, m.keys.SectionUp):
		m.sectionIdx = max(m.sectionIdx-1, 0)
		m.refreshNotesAndSelection()
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.focus == focusSidebar {
			m.noteList.CursorDown()
			m.syncSelection()
			return m, nil
		}
		m.preview.LineDown(1)
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.focus == focusSidebar {
			m.noteList.CursorUp()
			m.syncSelection()
			return m, nil
		}
		m.preview.LineUp(1)
		return m, nil

	case key.Matches(msg, m.keys.New):
		if sections[m.sectionIdx].key == fs.SectionTrash {
			return m, nil
		}
		note, err := m.store.Create(sections[m.sectionIdx].key)
		if err != nil {
			m.status = "create error: " + err.Error()
			return m, nil
		}

		m.refreshNotesAndReselect(note.ID)
		return m.startEditingSelected()

	case key.Matches(msg, m.keys.Edit):
		if sections[m.sectionIdx].key == fs.SectionTrash {
			return m, nil
		}
		return m.startEditingSelected()

	case key.Matches(msg, m.keys.Trash):
		if m.selected == nil {
			return m, nil
		}
		if sections[m.sectionIdx].key == fs.SectionTrash {
			if err := m.store.DeleteFromTrash(*m.selected); err != nil {
				m.status = "delete error: " + err.Error()
				return m, nil
			}

			m.status = "Deleted permanently: " + m.selected.Title
			m.refreshNotesAndSelection()
			return m, nil
		}

		updated, err := m.store.MoveToTrash(*m.selected)
		if err != nil {
			m.status = "trash error: " + err.Error()
			return m, nil
		}

		m.status = "Moved to Trash: " + updated.Title
		m.refreshNotesAndSelection()
		return m, nil

	case key.Matches(msg, m.keys.Restore):
		if m.selected == nil {
			return m, nil
		}
		if sections[m.sectionIdx].key != fs.SectionTrash {
			m.status = "restore works only in Trash"
			return m, nil
		}

		updated, err := m.store.RestoreFromTrash(*m.selected, fs.SectionNotes)
		if err != nil {
			m.status = "restore error: " + err.Error()
			return m, nil
		}

		m.status = "Restored: " + updated.Title
		m.refreshNotesAndSelection()
		return m, nil
	}

	var cmd tea.Cmd
	m.noteList, cmd = m.noteList.Update(msg)
	return m, cmd
}

func (m Model) renderSidebar() string {
	sec := sections[m.sectionIdx]
	secLine := titleStyle.Render("tenote") + " " + blurStyle.Render("•") + " " + focusStyle.Render(sec.title)
	if m.focus != focusSidebar {
		secLine = titleStyle.Render("tenote") + " " + blurStyle.Render("•") + " " + blurStyle.Render(sec.title)
	}

	box := border.Width(m.noteList.Width()).Height(m.noteList.Height()+2).Padding(0, 1)

	listView := m.noteList.View()
	return box.Render(secLine + "\n\n" + listView)
}

func (m Model) renderPreview() string {
	header := titleStyle.Render("Preview")
	content := m.preview.View()
	meta := m.renderPreviewMeta()

	if m.mode == modeEdit {
		header = titleStyle.Render("Edit")
		content = m.editor.View()
	} else {
		if m.previewErr != nil {
			content = "Error: " + m.previewErr.Error()
		}
		if strings.TrimSpace(content) == "" {
			content = blurStyle.Render("Select a note or press 'n' to create one.")
		}
	}

	w := m.width - m.noteList.Width() - 4
	if w < 20 {
		w = 20
	}

	box := border.Width(w).Height(m.noteList.Height()+2).Padding(0, 1)
	if m.focus == focusPreview {
		box = box.BorderForeground(lipgloss.Color("#25b067"))
	}
	return box.Render(header + "\n" + meta + "\n\n" + content)
}

func (m Model) renderPreviewMeta() string {
	noteTitle := "-"
	noteDate := "-"
	if m.selected != nil {
		noteTitle = m.selected.Title
		noteDate = m.selected.UpdatedAt.Format(timeLayout)
	}

	return strings.Join([]string{
		"---",
		"Note title: " + noteTitle,
		"Date: " + noteDate,
		"---",
	}, "\n")
}

func (m Model) renderStatus() string {
	if strings.TrimSpace(m.status) == "" {
		return ""
	}

	return lipgloss.NewStyle().Padding(0, 1).Render(statusStyle.Render(m.status))
}

func (m Model) renderHelp() string {
	if m.mode == modeEdit {
		return lipgloss.NewStyle().Padding(0, 1).Render(
			m.help.View(editKeyMap{KeyMap: m.keys}),
		)
	}
	if sections[m.sectionIdx].key == fs.SectionTrash {
		return lipgloss.NewStyle().Padding(0, 1).Render(
			m.help.View(trashKeyMap{KeyMap: m.keys}),
		)
	}

	return lipgloss.NewStyle().Padding(0, 1).Render(
		m.help.View(m.keys),
	)
}

// ---------- helpers ----------

func (m *Model) layout() {
	sidebarW := max(28, min(44, m.width/3))
	contentH := max(10, m.height-3)

	rightW := m.width - sidebarW - 6
	rightInnerH := contentH - 4
	previewBodyH := rightInnerH - 6
	if previewBodyH < 3 {
		previewBodyH = 3
	}

	m.noteList.SetSize(sidebarW-4, contentH-4)

	m.preview = viewport.New(rightW, previewBodyH)
	m.editor.SetWidth(rightW)
	m.editor.SetHeight(previewBodyH)

	wrapWidth := rightW - 2
	if wrapWidth < 20 {
		wrapWidth = 20
	}
	if r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(wrapWidth),
	); err == nil {
		m.renderer = r
	}

	m.syncSelection()
}

func (m *Model) refreshNotesAndSelection() {
	if err := m.reloadNotes(); err != nil {
		m.status = "load error: " + err.Error()
		return
	}

	m.syncSelection()
}

func (m *Model) refreshNotesAndReselect(noteID string) {
	if err := m.reloadNotes(); err != nil {
		m.status = "load error: " + err.Error()
		return
	}

	m.reselectByID(noteID)
	m.syncSelection()
}

func (m *Model) reloadNotes() error {
	sec := sections[m.sectionIdx].key
	notes, err := m.store.List(sec)
	if err != nil {
		return err
	}
	m.notes = notes

	items := make([]list.Item, 0, len(notes))
	for _, n := range notes {
		items = append(items, noteItem{n: n})
	}
	m.noteList.SetItems(items)
	m.noteList.Title = ""
	return nil
}

func (m *Model) syncSelection() {
	if m.mode == modeEdit {
		return
	}

	if len(m.notes) == 0 || len(m.noteList.Items()) == 0 {
		m.selected = nil
		m.preview.SetContent("")
		return
	}

	idx := m.noteList.Index()
	if idx < 0 || idx >= len(m.notes) {
		idx = 0
		m.noteList.Select(idx)
	}

	n := m.notes[idx]
	m.selected = &n

	body, err := m.store.ReadBody(n.Path)
	m.previewErr = err
	if err == nil {
		m.preview.SetContent(m.renderMarkdown(body))
	}
}

func (m *Model) reselectByID(id string) {
	for i, it := range m.noteList.Items() {
		ni, ok := it.(noteItem)
		if !ok {
			continue
		}
		if ni.n.ID == id {
			m.noteList.Select(i)
			return
		}
	}
}

func (m *Model) startEditingSelected() (Model, tea.Cmd) {
	if m.selected == nil {
		return *m, nil
	}

	body, err := m.store.ReadBody(m.selected.Path)
	if err != nil {
		m.status = "read error: " + err.Error()
		return *m, nil
	}

	m.mode = modeEdit
	m.dirty = false
	m.editErr = nil
	m.editor.SetValue(body)
	m.editor.CursorEnd()
	m.editor.Focus()
	m.focus = focusPreview

	return *m, nil
}

func (m *Model) renderMarkdown(body string) string {
	if m.renderer == nil {
		return body
	}
	rendered, err := m.renderer.Render(body)
	if err != nil {
		return body
	}
	return rendered
}

func (m *Model) exitEditMode(status string) {
	m.mode = modeBrowse
	m.editor.Blur()
	m.status = status
	m.dirty = false
	m.editErr = nil
	m.syncSelection()
}

type editKeyMap struct{ KeyMap }

func (k editKeyMap) ShortHelp() []key.Binding { return k.KeyMap.EditShortHelp() }

type trashKeyMap struct{ KeyMap }

func (k trashKeyMap) ShortHelp() []key.Binding { return k.KeyMap.TrashShortHelp() }
