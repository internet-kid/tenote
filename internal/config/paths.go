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

// ResolvePaths resolves and creates Tenote data directories using the saved config.
func ResolvePaths() (Paths, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return Paths{}, fmt.Errorf("load config: %w", err)
	}
	return ResolvePathsFrom(cfg.StorageDir)
}

// ResolvePathsFrom resolves and creates Tenote data directories rooted at root.
func ResolvePathsFrom(root string) (Paths, error) {
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
