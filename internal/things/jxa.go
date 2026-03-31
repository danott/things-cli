package things

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const defaultTimeout = 10 * time.Second

func RunJXA(script string) (json.RawMessage, error) {
	return RunJXAWithTimeout(script, defaultTimeout)
}

func RunJXAWithTimeout(script string, timeout time.Duration) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "osascript", "-l", "JavaScript")
	cmd.Stdin = strings.NewReader(script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("JXA script timed out after %s", timeout)
		}
		return nil, fmt.Errorf("JXA: %s: %w", strings.TrimSpace(stderr.String()), err)
	}

	raw := bytes.TrimSpace(stdout.Bytes())
	if len(raw) == 0 {
		return json.RawMessage("null"), nil
	}
	return json.RawMessage(raw), nil
}

// RunAppleScript executes an AppleScript string and returns the output.
func RunAppleScript(script string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("AppleScript timed out after %s", defaultTimeout)
		}
		return "", fmt.Errorf("AppleScript: %s: %w", strings.TrimSpace(stderr.String()), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// ListTodos returns todos from a named Things list.
func ListTodos(listName string) ([]Todo, error) {
	script := fmt.Sprintf(listTodosScript, escapeJS(listName))
	raw, err := RunJXA(script)
	if err != nil {
		return nil, fmt.Errorf("list todos from %q: %w", listName, err)
	}
	var todos []Todo
	if err := json.Unmarshal(raw, &todos); err != nil {
		return nil, fmt.Errorf("parse todos: %w", err)
	}
	return todos, nil
}

// ListProjectTodos returns todos belonging to a project.
func ListProjectTodos(projectName string) ([]Todo, error) {
	script := fmt.Sprintf(projectTodosScript, escapeJS(projectName), escapeJS(projectName))
	raw, err := RunJXA(script)
	if err != nil {
		return nil, fmt.Errorf("list project todos: %w", err)
	}
	var todos []Todo
	if err := json.Unmarshal(raw, &todos); err != nil {
		return nil, fmt.Errorf("parse todos: %w", err)
	}
	return todos, nil
}

// ListAreaTodos returns todos belonging to an area.
func ListAreaTodos(areaName string) ([]Todo, error) {
	script := fmt.Sprintf(areaTodosScript, escapeJS(areaName), escapeJS(areaName))
	raw, err := RunJXA(script)
	if err != nil {
		return nil, fmt.Errorf("list area todos: %w", err)
	}
	var todos []Todo
	if err := json.Unmarshal(raw, &todos); err != nil {
		return nil, fmt.Errorf("parse todos: %w", err)
	}
	return todos, nil
}

// ListTagTodos returns todos with a specific tag.
func ListTagTodos(tagName string) ([]Todo, error) {
	script := fmt.Sprintf(tagTodosScript, escapeJS(tagName))
	raw, err := RunJXA(script)
	if err != nil {
		return nil, fmt.Errorf("list tag todos: %w", err)
	}
	var todos []Todo
	if err := json.Unmarshal(raw, &todos); err != nil {
		return nil, fmt.Errorf("parse todos: %w", err)
	}
	return todos, nil
}

// GetTodo returns a single todo by ID.
func GetTodo(id string) (*Todo, error) {
	script := fmt.Sprintf(getTodoScript, escapeJS(id))
	raw, err := RunJXA(script)
	if err != nil {
		return nil, fmt.Errorf("get todo %q: %w", id, err)
	}
	var todo Todo
	if err := json.Unmarshal(raw, &todo); err != nil {
		return nil, fmt.Errorf("parse todo: %w", err)
	}
	return &todo, nil
}

// ListProjects returns all projects.
func ListProjects() ([]Project, error) {
	raw, err := RunJXA(listProjectsScript)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	var projects []Project
	if err := json.Unmarshal(raw, &projects); err != nil {
		return nil, fmt.Errorf("parse projects: %w", err)
	}
	return projects, nil
}

// GetProject returns a single project by ID.
func GetProject(id string) (*Project, error) {
	script := fmt.Sprintf(getProjectScript, escapeJS(id))
	raw, err := RunJXA(script)
	if err != nil {
		return nil, fmt.Errorf("get project %q: %w", id, err)
	}
	var project Project
	if err := json.Unmarshal(raw, &project); err != nil {
		return nil, fmt.Errorf("parse project: %w", err)
	}
	return &project, nil
}

// ListTags returns all tags.
func ListTags() ([]Tag, error) {
	raw, err := RunJXA(listTagsScript)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	var tags []Tag
	if err := json.Unmarshal(raw, &tags); err != nil {
		return nil, fmt.Errorf("parse tags: %w", err)
	}
	return tags, nil
}

// CreateTag creates a new tag via AppleScript.
func CreateTag(name string, parentTag string) error {
	if parentTag != "" {
		script := fmt.Sprintf(`tell application "Things3" to make new tag with properties {name:"%s", parent tag:tag "%s"}`,
			escapeJS(name), escapeJS(parentTag))
		_, err := RunAppleScript(script)
		return err
	}
	script := fmt.Sprintf(`tell application "Things3" to make new tag with properties {name:"%s"}`, escapeJS(name))
	_, err := RunAppleScript(script)
	return err
}

// RenameTag renames a tag via AppleScript.
func RenameTag(oldName, newName string) error {
	script := fmt.Sprintf(`tell application "Things3" to set name of tag "%s" to "%s"`,
		escapeJS(oldName), escapeJS(newName))
	_, err := RunAppleScript(script)
	return err
}

// DeleteTag deletes a tag via AppleScript.
func DeleteTag(name string) error {
	script := fmt.Sprintf(`tell application "Things3" to delete tag "%s"`, escapeJS(name))
	_, err := RunAppleScript(script)
	return err
}

// CreateArea creates a new area via AppleScript.
func CreateArea(name string) error {
	script := fmt.Sprintf(`tell application "Things3" to make new area with properties {name:"%s"}`, escapeJS(name))
	_, err := RunAppleScript(script)
	return err
}

// RenameArea renames an area via AppleScript.
func RenameArea(oldName, newName string) error {
	script := fmt.Sprintf(`tell application "Things3" to set name of area "%s" to "%s"`,
		escapeJS(oldName), escapeJS(newName))
	_, err := RunAppleScript(script)
	return err
}

// DeleteArea deletes an area via AppleScript.
func DeleteArea(name string) error {
	script := fmt.Sprintf(`tell application "Things3" to delete area "%s"`, escapeJS(name))
	_, err := RunAppleScript(script)
	return err
}

// CompleteProject marks a project as completed via AppleScript.
func CompleteProject(id string) error {
	script := fmt.Sprintf(`tell application "Things3" to set status of project id "%s" to completed`, escapeJS(id))
	_, err := RunAppleScript(script)
	return err
}

// CancelProject marks a project as canceled via AppleScript.
func CancelProject(id string) error {
	script := fmt.Sprintf(`tell application "Things3" to set status of project id "%s" to canceled`, escapeJS(id))
	_, err := RunAppleScript(script)
	return err
}

// DeleteProject moves a project to Trash via AppleScript.
func DeleteProject(id string) error {
	script := fmt.Sprintf(`tell application "Things3" to delete project id "%s"`, escapeJS(id))
	_, err := RunAppleScript(script)
	return err
}

// CompleteTodo marks a todo as completed via AppleScript.
func CompleteTodo(id string) error {
	script := fmt.Sprintf(`tell application "Things3" to set status of to do id "%s" to completed`, escapeJS(id))
	_, err := RunAppleScript(script)
	return err
}

// CancelTodo marks a todo as canceled via AppleScript.
func CancelTodo(id string) error {
	script := fmt.Sprintf(`tell application "Things3" to set status of to do id "%s" to canceled`, escapeJS(id))
	_, err := RunAppleScript(script)
	return err
}

// DeleteTodo moves a todo to Trash via AppleScript.
func DeleteTodo(id string) error {
	script := fmt.Sprintf(`tell application "Things3" to delete to do id "%s"`, escapeJS(id))
	_, err := RunAppleScript(script)
	return err
}

// LogCompleted triggers Things' "Log Completed" menu action via System Events.
// This sets leavesTombstone=1 on completed items, moving them to the Logbook.
// It does not steal focus from the current application.
func LogCompleted() error {
	script := `tell application "System Events" to tell process "Things3" to click menu item "Log Completed" of menu "Items" of menu bar 1`
	_, err := RunAppleScript(script)
	return err
}

// GetThingsVersion returns the Things app version.
func GetThingsVersion() (string, error) {
	return RunAppleScript(`tell application "Things3" to return version`)
}

func escapeJS(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}
