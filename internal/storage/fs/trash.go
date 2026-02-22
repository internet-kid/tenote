package fs

import (
	"fmt"
	"os"
	"time"
)

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
