package config

import (
	"os"
	"path/filepath"
)

type Paths struct {
	Root  string
	Notes string
	Todo  string
	Trash string
}

func ResolvePaths() (Paths, error) {
	// XDG-ish: ~/.local/share/tenote
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	root := filepath.Join(home, ".local", "share", "tenote")
	p := Paths{
		Root:  root,
		Notes: filepath.Join(root, "notes"),
		Todo:  filepath.Join(root, "todo"),
		Trash: filepath.Join(root, "trash"),
	}

	for _, dir := range []string{p.Root, p.Notes, p.Todo, p.Trash} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return Paths{}, err
		}
	}

	return p, nil
}
