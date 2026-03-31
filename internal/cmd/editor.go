package cmd

import "github.com/danott/things-cli/internal/editor"

// openInEditor writes content to a temp file, opens it in $VISUAL or $EDITOR,
// waits for the editor to exit, and returns the saved contents.
func openInEditor(content string) (string, error) {
	return editor.Open(content)
}
