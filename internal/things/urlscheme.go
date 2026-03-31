package things

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

func StatusPtr(s Status) *Status { return &s }

type URLBuilder struct {
	command string
	params  url.Values
}

func NewURL(command string) *URLBuilder {
	return &URLBuilder{
		command: command,
		params:  url.Values{},
	}
}

func (b *URLBuilder) Set(key, value string) *URLBuilder {
	if value != "" {
		b.params.Set(key, value)
	}
	return b
}

func (b *URLBuilder) SetBool(key string, value bool) *URLBuilder {
	if value {
		b.params.Set(key, "true")
	}
	return b
}

// SetStatus translates a *Status to the URL scheme's completed/canceled params.
// nil means no change; StatusOpen sends completed=false to mark incomplete.
func (b *URLBuilder) SetStatus(s *Status) *URLBuilder {
	if s == nil {
		return b
	}
	switch *s {
	case StatusCompleted:
		b.params.Set("completed", "true")
	case StatusCanceled:
		b.params.Set("canceled", "true")
	case StatusOpen:
		b.params.Set("completed", "false")
	}
	return b
}

func (b *URLBuilder) Build() string {
	u := fmt.Sprintf("things:///%s", b.command)
	if encoded := b.params.Encode(); encoded != "" {
		u += "?" + strings.ReplaceAll(encoded, "+", "%20")
	}
	return u
}

func (b *URLBuilder) Open() error {
	return OpenURL(b.Build())
}

func OpenURL(thingsURL string) error {
	cmd := exec.Command("open", "-g", thingsURL)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("open URL: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// AddTodo builds a things:///add URL.
func AddTodo(title string, opts AddTodoOptions) *URLBuilder {
	b := NewURL("add")
	b.Set("title", title)
	b.Set("notes", opts.Notes)
	b.Set("when", opts.When)
	b.Set("deadline", opts.Deadline)
	b.Set("tags", opts.Tags)
	b.Set("checklist-items", opts.ChecklistItems)
	b.Set("list", opts.List)
	b.Set("list-id", opts.ListID)
	b.Set("heading", opts.Heading)
	b.Set("heading-id", opts.HeadingID)
	b.SetBool("reveal", opts.Reveal)
	b.SetBool("show-quick-entry", opts.ShowQuickEntry)
	return b
}

type AddTodoOptions struct {
	Notes          string
	When           string
	Deadline       string
	Tags           string
	ChecklistItems string
	List           string
	ListID         string
	Heading        string
	HeadingID      string
	Reveal         bool
	ShowQuickEntry bool
}

// UpdateTodo builds a things:///update URL.
func UpdateTodo(id, authToken string, opts UpdateTodoOptions) *URLBuilder {
	b := NewURL("update")
	b.Set("id", id)
	b.Set("auth-token", authToken)
	b.Set("title", opts.Title)
	b.Set("notes", opts.Notes)
	b.Set("prepend-notes", opts.PrependNotes)
	b.Set("append-notes", opts.AppendNotes)
	b.Set("when", opts.When)
	b.Set("deadline", opts.Deadline)
	b.Set("tags", opts.Tags)
	b.Set("add-tags", opts.AddTags)
	b.Set("checklist-items", opts.ChecklistItems)
	b.Set("prepend-checklist-items", opts.PrependChecklistItems)
	b.Set("append-checklist-items", opts.AppendChecklistItems)
	b.Set("list", opts.List)
	b.Set("list-id", opts.ListID)
	b.Set("heading", opts.Heading)
	b.Set("heading-id", opts.HeadingID)
	b.SetStatus(opts.Status)
	b.SetBool("reveal", opts.Reveal)
	b.SetBool("duplicate", opts.Duplicate)
	return b
}

type UpdateTodoOptions struct {
	Title                 string
	Notes                 string
	PrependNotes          string
	AppendNotes           string
	When                  string
	Deadline              string
	Tags                  string
	AddTags               string
	ChecklistItems        string
	PrependChecklistItems string
	AppendChecklistItems  string
	List                  string
	ListID                string
	Heading               string
	HeadingID             string
	Status                *Status
	Reveal                bool
	Duplicate             bool
}

// UpdateProject builds a things:///update-project URL.
func UpdateProject(id, authToken string, opts UpdateProjectOptions) *URLBuilder {
	b := NewURL("update-project")
	b.Set("id", id)
	b.Set("auth-token", authToken)
	b.Set("title", opts.Title)
	b.Set("notes", opts.Notes)
	b.Set("prepend-notes", opts.PrependNotes)
	b.Set("append-notes", opts.AppendNotes)
	b.Set("when", opts.When)
	b.Set("deadline", opts.Deadline)
	b.Set("tags", opts.Tags)
	b.Set("add-tags", opts.AddTags)
	b.Set("area", opts.Area)
	b.Set("area-id", opts.AreaID)
	b.SetStatus(opts.Status)
	b.SetBool("reveal", opts.Reveal)
	return b
}

type UpdateProjectOptions struct {
	Title        string
	Notes        string
	PrependNotes string
	AppendNotes  string
	When         string
	Deadline     string
	Tags         string
	AddTags      string
	Area         string
	AreaID       string
	Status       *Status
	Reveal       bool
}

// AddProject builds a things:///add-project URL.
func AddProject(title string, opts AddProjectOptions) *URLBuilder {
	b := NewURL("add-project")
	b.Set("title", title)
	b.Set("notes", opts.Notes)
	b.Set("when", opts.When)
	b.Set("deadline", opts.Deadline)
	b.Set("tags", opts.Tags)
	b.Set("area", opts.Area)
	b.Set("area-id", opts.AreaID)
	b.Set("to-dos", opts.Todos)
	b.SetBool("reveal", opts.Reveal)
	return b
}

type AddProjectOptions struct {
	Notes    string
	When     string
	Deadline string
	Tags     string
	Area     string
	AreaID   string
	Todos    string
	Reveal   bool
}

// ShowItem builds a things:///show URL.
func ShowItem(id string) *URLBuilder {
	b := NewURL("show")
	b.Set("id", id)
	return b
}

// ShowList builds a things:///show URL for a built-in list.
func ShowList(listID string) *URLBuilder {
	b := NewURL("show")
	b.Set("id", listID)
	return b
}

// Search builds a things:///search URL.
func Search(query string) *URLBuilder {
	b := NewURL("search")
	b.Set("query", query)
	return b
}

// JSONCommand builds a things:///json URL.
func JSONCommand(data, authToken string, reveal bool) *URLBuilder {
	b := NewURL("json")
	b.Set("data", data)
	b.Set("auth-token", authToken)
	b.SetBool("reveal", reveal)
	return b
}
