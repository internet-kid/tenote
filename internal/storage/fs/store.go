package fs

import (
	"path/filepath"

	"github.com/internet-kid/tenote/internal/config"
)

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
	case SectionTrash:
		return s.paths.Trash
	default:
		return s.paths.Notes
	}
}

func (s *Store) notePath(section Section, id string) string {
	return filepath.Join(s.dirFor(section), id+noteExt)
}
