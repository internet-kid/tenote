package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	dirPerm = 0o755
)

type Paths struct {
	Root  string
	Notes string
	Trash string
}

// ResolvePaths resolves and creates Tenote data directories.
func ResolvePaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("resolve user home dir: %w", err)
	}

	root := filepath.Join(home, ".local", "share", "tenote")
	p := Paths{
		Root:  root,
		Notes: filepath.Join(root, "notes"),
		Trash: filepath.Join(root, "trash"),
	}

	for _, dir := range []string{p.Root, p.Notes, p.Trash} {
		if err := os.MkdirAll(dir, dirPerm); err != nil {
			return Paths{}, fmt.Errorf("create data dir %q: %w", dir, err)
		}
	}

	return p, nil
}
