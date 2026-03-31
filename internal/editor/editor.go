package editor

import (
	"fmt"
	"os"
	"os/exec"
)

// Command returns the user's preferred editor command name.
func Command() string {
	if e := os.Getenv("VISUAL"); e != "" {
		return e
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "vi"
}

// Open writes content to a temp file, opens it in the editor,
// waits for the editor to exit, and returns the saved contents.
func Open(content string) (string, error) {
	f, err := os.CreateTemp("", "things-*.md")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(content); err != nil {
		f.Close()
		return "", fmt.Errorf("write temp file: %w", err)
	}
	f.Close()

	cmd := exec.Command(Command(), f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor: %w", err)
	}

	result, err := os.ReadFile(f.Name())
	if err != nil {
		return "", fmt.Errorf("read temp file: %w", err)
	}
	return string(result), nil
}

// TempFile creates a temp file with the given content and returns its path.
// The caller is responsible for removing the file.
func TempFile(content string) (string, error) {
	f, err := os.CreateTemp("", "things-*.md")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", fmt.Errorf("write temp file: %w", err)
	}
	f.Close()
	return f.Name(), nil
}
