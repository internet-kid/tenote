package fs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/oklog/ulid/v2"
)

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

func (s *Store) WriteBody(path, body string) error {
	if err := os.WriteFile(path, []byte(body), filePerm); err != nil {
		return fmt.Errorf("write note %q: %w", path, err)
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
		// strip markdown heading prefix
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
