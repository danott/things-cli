package things

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParsedChecklistItem holds a checklist item with its completion state as
// parsed from markdown. Used when the JSON command is needed to preserve
// completion status on write.
type ParsedChecklistItem struct {
	Title     string
	Completed bool
}

// TodoToMarkdown renders a Todo as an editable markdown document.
func TodoToMarkdown(t *Todo) string {
	var b strings.Builder

	// Frontmatter — always emit all fields so the editor shows what's editable
	when := ""
	if t.ActivationDate != nil {
		when = t.ActivationDate.Format("2006-01-02")
	}
	deadline := ""
	if t.DueDate != nil {
		deadline = t.DueDate.Format("2006-01-02")
	}
	list := t.ProjectName
	if list == "" {
		list = t.AreaName
	}
	fmt.Fprintf(&b, "---\nstatus: %s\nwhen: %s\ndeadline: %s\ntags: %s\nlist: %s\n---\n\n",
		t.Status, when, deadline, t.TagNames, list)

	// Title as H1
	fmt.Fprintf(&b, "# %s\n", t.Name)

	// Notes
	if t.Notes != "" {
		fmt.Fprintf(&b, "\n%s\n", t.Notes)
	}

	// Checklist items — [x] for completed, [ ] for open
	// Note: when writing back via URL scheme, completion status cannot be
	// preserved per-item; all items will be reset to open.
	if len(t.ChecklistItems) > 0 {
		b.WriteString("\n")
		for _, item := range t.ChecklistItems {
			if item.Status == StatusCompleted {
				fmt.Fprintf(&b, "- [x] %s\n", item.Name)
			} else {
				fmt.Fprintf(&b, "- [ ] %s\n", item.Name)
			}
		}
	}

	return b.String()
}

// ProjectToMarkdown renders a Project as a markdown document with frontmatter,
// matching the format used by TodoToMarkdown.
func ProjectToMarkdown(p *Project) string {
	var b strings.Builder

	when := ""
	if p.ActivationDate != nil {
		when = p.ActivationDate.Format("2006-01-02")
	}
	deadline := ""
	if p.DueDate != nil {
		deadline = p.DueDate.Format("2006-01-02")
	}
	fmt.Fprintf(&b, "---\nstatus: %s\nwhen: %s\ndeadline: %s\ntags: %s\narea: %s\n---\n\n",
		p.Status, when, deadline, p.TagNames, p.AreaName)

	fmt.Fprintf(&b, "# %s\n", p.Name)
	if p.Notes != "" {
		fmt.Fprintf(&b, "\n%s\n", p.Notes)
	}
	return b.String()
}

// BlankTodoMarkdown returns a template for creating a new todo in the editor.
func BlankTodoMarkdown() string {
	return "---\nwhen: \ndeadline: \ntags: \nlist: \n---\n\n# \n"
}

// ParseTodoMarkdown parses an editor markdown document into a title and fields.
// It returns the title and the notes/checklist/metadata extracted from the document.
func ParseTodoMarkdown(src string) (title string, opts AddTodoOptions, err error) {
	fm, body := SplitFrontmatter(src)
	if fm != "" {
		parseFrontmatterIntoAdd(fm, &opts)
	}
	title, notes, items := parseBody(body)
	opts.Notes = notes
	opts.ChecklistItems = ChecklistItemsString(items)
	return title, opts, nil
}

// ParseTodoMarkdownForUpdate parses an editor markdown document into update options.
func ParseTodoMarkdownForUpdate(src string) (title string, opts UpdateTodoOptions, checklist []ParsedChecklistItem, err error) {
	fm, body := SplitFrontmatter(src)
	if fm != "" {
		parseFrontmatterIntoUpdate(fm, &opts)
	}
	title, notes, checklist := parseBody(body)
	opts.Notes = notes
	opts.ChecklistItems = ChecklistItemsString(checklist)
	opts.Title = title
	return title, opts, checklist, nil
}

// SplitFrontmatter splits a document into frontmatter and body.
// Frontmatter is the content between the first two --- delimiters.
func SplitFrontmatter(src string) (frontmatter, body string) {
	if !strings.HasPrefix(src, "---\n") {
		return "", src
	}
	rest := src[4:] // skip opening ---\n
	end := strings.Index(rest, "\n---\n")
	if end == -1 {
		return "", src
	}
	return rest[:end], rest[end+5:] // skip \n---\n
}

func parseFrontmatterIntoAdd(fm string, opts *AddTodoOptions) {
	for _, line := range strings.Split(fm, "\n") {
		key, val, ok := parseFrontmatterLine(line)
		if !ok {
			continue
		}
		switch key {
		case "when":
			opts.When = val
		case "deadline":
			opts.Deadline = val
		case "tags":
			opts.Tags = val
		case "list":
			opts.List = val
		case "list-id":
			opts.ListID = val
		case "heading":
			opts.Heading = val
		}
	}
}

func parseFrontmatterIntoUpdate(fm string, opts *UpdateTodoOptions) {
	for _, line := range strings.Split(fm, "\n") {
		key, val, ok := parseFrontmatterLine(line)
		if !ok {
			continue
		}
		switch key {
		case "status":
			switch Status(val) {
			case StatusCompleted:
				opts.Status = StatusPtr(StatusCompleted)
			case StatusCanceled:
				opts.Status = StatusPtr(StatusCanceled)
			default:
				// "open" or any other value — mark as incomplete.
				// completed=false works regardless of prior state (completed or canceled).
				opts.Status = StatusPtr(StatusOpen)
			}
		case "when":
			opts.When = val
		case "deadline":
			opts.Deadline = val
		case "tags":
			opts.Tags = val
		case "list":
			opts.List = val
		case "list-id":
			opts.ListID = val
		case "heading":
			opts.Heading = val
		}
	}
}

func parseFrontmatterLine(line string) (key, val string, ok bool) {
	i := strings.IndexByte(line, ':')
	if i == -1 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:i])
	val = strings.TrimSpace(line[i+1:])
	return key, val, key != ""
}

// ParsedChecklistItemsFromMarkdown returns structured checklist items parsed
// from a markdown document, preserving completion state.
func ParsedChecklistItemsFromMarkdown(src string) []ParsedChecklistItem {
	_, body := SplitFrontmatter(src)
	_, _, items := parseBody(body)
	return items
}

// ChecklistItemsString converts parsed checklist items to the newline-separated
// string expected by the URL scheme (completion state is discarded).
func ChecklistItemsString(items []ParsedChecklistItem) string {
	titles := make([]string, len(items))
	for i, item := range items {
		titles[i] = item.Title
	}
	return strings.Join(titles, "\n")
}

// BuildTodoUpdateJSON builds a Things JSON command payload for a full todo
// update, including all metadata fields and checklist items with per-item
// completion state. Use this instead of the URL scheme update when checklist
// completion state must be preserved.
func BuildTodoUpdateJSON(id string, opts UpdateTodoOptions, items []ParsedChecklistItem) (string, error) {
	type checklistItemAttrs struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed,omitempty"`
	}
	type checklistItemObj struct {
		Type       string             `json:"type"`
		Attributes checklistItemAttrs `json:"attributes"`
	}
	type todoAttrs struct {
		Title          string             `json:"title,omitempty"`
		Notes          string             `json:"notes,omitempty"`
		When           string             `json:"when,omitempty"`
		Deadline       string             `json:"deadline,omitempty"`
		Tags           []string           `json:"tags,omitempty"`
		List           string             `json:"list,omitempty"`
		ListID         string             `json:"list-id,omitempty"`
		Heading        string             `json:"heading,omitempty"`
		HeadingID      string             `json:"heading-id,omitempty"`
		Completed      *bool `json:"completed,omitempty"`
		Canceled       *bool `json:"canceled,omitempty"`
		ChecklistItems []checklistItemObj `json:"checklist-items,omitempty"`
	}
	type todoObj struct {
		Type       string    `json:"type"`
		Operation  string    `json:"operation"`
		ID         string    `json:"id"`
		Attributes todoAttrs `json:"attributes"`
	}

	attrs := todoAttrs{
		Title:     opts.Title,
		Notes:     opts.Notes,
		When:      opts.When,
		Deadline:  opts.Deadline,
		List:      opts.List,
		ListID:    opts.ListID,
		Heading:   opts.Heading,
		HeadingID: opts.HeadingID,
	}
	if opts.Status != nil {
		t, f := true, false
		switch *opts.Status {
		case StatusCompleted:
			attrs.Completed = &t
		case StatusCanceled:
			attrs.Canceled = &t
		case StatusOpen:
			attrs.Completed = &f
		}
	}

	for _, t := range strings.Split(opts.Tags, ",") {
		if tag := strings.TrimSpace(t); tag != "" {
			attrs.Tags = append(attrs.Tags, tag)
		}
	}

	for _, item := range items {
		attrs.ChecklistItems = append(attrs.ChecklistItems, checklistItemObj{
			Type:       "checklist-item",
			Attributes: checklistItemAttrs{Title: item.Title, Completed: item.Completed},
		})
	}

	payload, err := json.Marshal([]todoObj{{
		Type:      "to-do",
		Operation: "update",
		ID:        id,
		Attributes: attrs,
	}})
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

// parseBody extracts the title (H1), notes, and checklist items from the body.
// Checklist items are lines starting with "- [ ]" or "- [x]".
// Everything else (excluding the H1 line) is notes.
func parseBody(body string) (title, notes string, checklist []ParsedChecklistItem) {
	var noteLines []string

	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "# ") && title == "" {
			title = strings.TrimPrefix(line, "# ")
			continue
		}
		if strings.HasPrefix(line, "- [ ] ") {
			checklist = append(checklist, ParsedChecklistItem{Title: strings.TrimPrefix(line, "- [ ] ")})
			continue
		}
		if strings.HasPrefix(line, "- [x] ") {
			checklist = append(checklist, ParsedChecklistItem{Title: strings.TrimPrefix(line, "- [x] "), Completed: true})
			continue
		}
		noteLines = append(noteLines, line)
	}

	notes = strings.TrimSpace(strings.Join(noteLines, "\n"))
	return title, notes, checklist
}
