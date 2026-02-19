package fs

import (
	"bufio"
	"errors"
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

type Store struct {
	paths config.Paths
}

func NewStore(paths config.Paths) *Store {
	return &Store{paths: paths}
}

func (s *Store) DirFor(section Section) string {
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
	dir := s.DirFor(section)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var notes []Note
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		path := filepath.Join(dir, name)

		info, err := e.Info()
		if err != nil {
			continue
		}

		title := readFirstNonEmptyLine(path)
		id := strings.TrimSuffix(name, ".md")
		if title == "" {
			title = id
		}

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
		return "", err
	}
	return string(b), nil
}

func (s *Store) Create(section Section) (Note, error) {
	id := ulid.Make().String()
	path := filepath.Join(s.DirFor(section), id+".md")

	template := "# " + id + "\n\n"

	if err := os.WriteFile(path, []byte(template), 0o644); err != nil {
		return Note{}, err
	}

	title := readFirstNonEmptyLine(path)
	if title == "" {
		title = id
	}

	info, _ := os.Stat(path)
	mod := time.Now()
	if info != nil {
		mod = info.ModTime()
	}

	return Note{
		ID:        id,
		Title:     title,
		Path:      path,
		Section:   section,
		UpdatedAt: mod,
	}, nil
}

func (s *Store) MoveToTrash(n Note) (Note, error) {
	if n.Section == SectionTrash {
		return n, nil
	}
	dst := filepath.Join(s.paths.Trash, n.ID+".md")
	if err := os.Rename(n.Path, dst); err != nil {
		return Note{}, err
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
	dst := filepath.Join(s.DirFor(target), n.ID+".md")
	if err := os.Rename(n.Path, dst); err != nil {
		return Note{}, err
	}
	n.Path = dst
	n.Section = target
	n.UpdatedAt = time.Now()
	return n, nil
}

func readFirstNonEmptyLine(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
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
		return line
	}
	return ""
}

func (s *Store) WriteBody(path, body string) error {
	return os.WriteFile(path, []byte(body), 0o644)
}

var ErrNotFound = errors.New("note not found")
