package editor

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

func EditorCommand() string {
	if v := os.Getenv("EDITOR"); v != "" {
		return v
	}
	return "vi"
}

func EditCmd(path string) (*exec.Cmd, error) {
	editor := strings.TrimSpace(EditorCommand())
	if editor == "" {
		return nil, errors.New("EDITOR is empty")
	}

	parts := strings.Fields(editor)
	name := parts[0]
	args := append(parts[1:], path)

	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}
