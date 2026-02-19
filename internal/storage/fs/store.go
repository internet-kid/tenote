package fs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"tenote/internal/config"
)

type Section string

const (
	SectionTodo  Section = "todo"
	SectionNotes Section = "notes"
	SectionTrash Section = "trash"
)

type Note struct {
	ID        string
	Title     string
	Path      string
	Section   Section
	UpdatedAt time.Time
}

const (
	filePerm        = 0o644
	noteExt         = ".md"
	noteTemplate    = "# \n\n"
	defaultNoteName = "(untitled)"
)

type Store struct {
	paths config.Paths
}

func NewStore(paths config.Paths) *Store {
	return &Store{paths: paths}
}

func (s *Store) dirFor(section Section) string {
	switch section {
	case SectionTodo:
		return s.paths.Todo
	case SectionTrash:
		return s.paths.Trash
	default:
		return s.paths.Notes
	}
}

func (s *Store) List(section Section) ([]Note, error) {
	dir := s.dirFor(section)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read section dir %q: %w", dir, err)
	}

	var notes []Note
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, noteExt) {
			continue
		}

		path := filepath.Join(dir, name)

		info, err := e.Info()
		if err != nil {
			return nil, fmt.Errorf("read file info %q: %w", path, err)
		}

		title, err := readFirstNonEmptyLine(path)
		if err != nil {
			return nil, err
		}
		if title == "" {
			title = defaultNoteName
		}

		id := strings.TrimSuffix(name, noteExt)

		notes = append(notes, Note{
			ID:        id,
			Title:     title,
			Path:      path,
			Section:   section,
			UpdatedAt: info.ModTime(),
		})
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].UpdatedAt.After(notes[j].UpdatedAt)
	})

	return notes, nil
}

func (s *Store) ReadBody(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read note %q: %w", path, err)
	}
	return string(b), nil
}

func (s *Store) Create(section Section) (Note, error) {
	id := ulid.Make().String()
	path := s.notePath(section, id)

	if err := os.WriteFile(path, []byte(noteTemplate), filePerm); err != nil {
		return Note{}, fmt.Errorf("create note %q: %w", path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return Note{}, fmt.Errorf("stat new note %q: %w", path, err)
	}

	return Note{
		ID:        id,
		Title:     defaultNoteName,
		Path:      path,
		Section:   section,
		UpdatedAt: info.ModTime(),
	}, nil
}

func (s *Store) MoveToTrash(n Note) (Note, error) {
	if n.Section == SectionTrash {
		return n, nil
	}
	dst := s.notePath(SectionTrash, n.ID)
	if err := os.Rename(n.Path, dst); err != nil {
		return Note{}, fmt.Errorf("move note %q to trash: %w", n.Path, err)
	}
	n.Path = dst
	n.Section = SectionTrash
	n.UpdatedAt = time.Now()
	return n, nil
}

func (s *Store) RestoreFromTrash(n Note, target Section) (Note, error) {
	if n.Section != SectionTrash {
		return n, nil
	}
	if target == SectionTrash {
		target = SectionNotes
	}
	dst := s.notePath(target, n.ID)
	if err := os.Rename(n.Path, dst); err != nil {
		return Note{}, fmt.Errorf("restore note %q: %w", n.Path, err)
	}
	n.Path = dst
	n.Section = target
	n.UpdatedAt = time.Now()
	return n, nil
}

func (s *Store) DeleteFromTrash(n Note) error {
	if n.Section != SectionTrash {
		return fmt.Errorf("delete from trash requires trash section, got %q", n.Section)
	}

	if err := os.Remove(n.Path); err != nil {
		return fmt.Errorf("delete note %q from trash: %w", n.Path, err)
	}

	return nil
}

func readFirstNonEmptyLine(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open note %q: %w", path, err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		// if markdown heading
		if strings.HasPrefix(line, "#") {
			line = strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
		return line, nil
	}

	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("scan note %q: %w", path, err)
	}

	return "", nil
}

func (s *Store) WriteBody(path, body string) error {
	if err := os.WriteFile(path, []byte(body), filePerm); err != nil {
		return fmt.Errorf("write note %q: %w", path, err)
	}

	return nil
}

func (s *Store) notePath(section Section, id string) string {
	return filepath.Join(s.dirFor(section), id+noteExt)
}
