package editor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const defaultEditor = "vi"

func EditorCommand() string {
	if v := os.Getenv("EDITOR"); v != "" {
		return v
	}
	return defaultEditor
}

func EditCmd(path string) (*exec.Cmd, error) {
	editor := strings.TrimSpace(EditorCommand())
	if editor == "" {
		return nil, errors.New("EDITOR is empty")
	}

	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return nil, errors.New("EDITOR has no command parts")
	}

	name := parts[0]
	args := append(parts[1:], path)

	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if cmd.Err != nil {
		return nil, fmt.Errorf("resolve editor command %q: %w", name, cmd.Err)
	}

	return cmd, nil
}
