package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/danott/things-cli/internal/things"
)

func PrintTodosText(w io.Writer, todos []things.Todo, verbose bool) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, t := range todos {
		check := "  "
		if t.Status == things.StatusCompleted {
			check = "x "
		} else if t.Status == things.StatusCanceled {
			check = "- "
		}
		extra := ""
		if t.TagNames != "" {
			extra = "  [" + t.TagNames + "]"
		}
		project := ""
		if t.ProjectName != "" {
			project = t.ProjectName
		} else if t.AreaName != "" {
			project = t.AreaName
		}
		if verbose {
			fmt.Fprintf(tw, "%s%s\t%s\t%s\t%s\n", check, t.Name, t.ID, project, extra)
		} else {
			fmt.Fprintf(tw, "%s%s\t%s\t%s\n", check, t.Name, project, extra)
		}
	}
	tw.Flush()
}

func PrintTodosMarkdown(w io.Writer, todos []things.Todo) {
	for _, t := range todos {
		check := " "
		if t.Status == things.StatusCompleted {
			check = "x"
		} else if t.Status == things.StatusCanceled {
			check = "-"
		}
		fmt.Fprintf(w, "- [%s] %s\n", check, t.Name)
	}
}

func PrintTodosJSON(w io.Writer, todos []things.Todo) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(todos)
}

func PrintTodoText(w io.Writer, t *things.Todo) {
	fmt.Fprint(w, things.TodoToMarkdown(t))
}

func PrintTodoMarkdown(w io.Writer, t *things.Todo) {
	check := " "
	if t.Status == things.StatusCompleted {
		check = "x"
	} else if t.Status == things.StatusCanceled {
		check = "-"
	}
	fmt.Fprintf(w, "- [%s] %s\n", check, t.Name)
	if t.ProjectName != "" {
		fmt.Fprintf(w, "  **Project:** %s\n", t.ProjectName)
	}
	if t.AreaName != "" {
		fmt.Fprintf(w, "  **Area:** %s\n", t.AreaName)
	}
	if t.TagNames != "" {
		fmt.Fprintf(w, "  **Tags:** %s\n", t.TagNames)
	}
	if t.DueDate != nil {
		fmt.Fprintf(w, "  **Deadline:** %s\n", t.DueDate.Format("2006-01-02"))
	}
	if t.ActivationDate != nil {
		fmt.Fprintf(w, "  **When:** %s\n", t.ActivationDate.Format("2006-01-02"))
	}
	if t.Notes != "" {
		fmt.Fprintf(w, "\n%s\n", t.Notes)
	}
	if len(t.ChecklistItems) > 0 {
		fmt.Fprintln(w)
		for _, item := range t.ChecklistItems {
			itemCheck := " "
			if item.Status == things.StatusCompleted {
				itemCheck = "x"
			}
			fmt.Fprintf(w, "  - [%s] %s\n", itemCheck, item.Name)
		}
	}
}

func PrintTodoJSON(w io.Writer, t *things.Todo) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(t)
}

func PrintProjectsText(w io.Writer, projects []things.Project) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, p := range projects {
		check := "  "
		if p.Status == things.StatusCompleted {
			check = "x "
		} else if p.Status == things.StatusCanceled {
			check = "- "
		}
		fmt.Fprintf(tw, "%s%s\t%s\n", check, p.Name, p.AreaName)
	}
	tw.Flush()
}

func PrintProjectsMarkdown(w io.Writer, projects []things.Project) {
	for _, p := range projects {
		check := " "
		if p.Status == things.StatusCompleted {
			check = "x"
		} else if p.Status == things.StatusCanceled {
			check = "-"
		}
		fmt.Fprintf(w, "- [%s] %s\n", check, p.Name)
	}
}

func PrintProjectsJSON(w io.Writer, projects []things.Project) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(projects)
}

func PrintProjectText(w io.Writer, p *things.Project) {
	fmt.Fprint(w, things.ProjectToMarkdown(p))
}

func PrintProjectMarkdown(w io.Writer, p *things.Project) {
	fmt.Fprintf(w, "# %s\n", p.Name)
	if p.Status != things.StatusOpen {
		fmt.Fprintf(w, "\n**Status:** %s\n", p.Status)
	}
	if p.AreaName != "" {
		fmt.Fprintf(w, "\n**Area:** %s\n", p.AreaName)
	}
	if p.Notes != "" {
		fmt.Fprintf(w, "\n%s\n", p.Notes)
	}
}

func PrintProjectJSON(w io.Writer, p *things.Project) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}

func PrintAreasText(w io.Writer, areas []things.Area) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, a := range areas {
		fmt.Fprintf(tw, "%s\t%s\n", a.Name, a.ID)
	}
	tw.Flush()
}

func PrintAreasMarkdown(w io.Writer, areas []things.Area) {
	for _, a := range areas {
		fmt.Fprintf(w, "- %s\n", a.Name)
	}
}

func PrintAreasJSON(w io.Writer, areas []things.Area) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(areas)
}

func PrintTagsText(w io.Writer, tags []things.Tag) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, t := range tags {
		parent := ""
		if t.ParentTag != "" {
			parent = "(" + t.ParentTag + ")"
		}
		fmt.Fprintf(tw, "%s\t%s\n", t.Name, parent)
	}
	tw.Flush()
}

func PrintTagsMarkdown(w io.Writer, tags []things.Tag) {
	for _, t := range tags {
		if t.ParentTag != "" {
			fmt.Fprintf(w, "- %s (%s)\n", t.Name, t.ParentTag)
		} else {
			fmt.Fprintf(w, "- %s\n", t.Name)
		}
	}
}

func PrintTagsJSON(w io.Writer, tags []things.Tag) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(tags)
}

// ViewNameToListName maps CLI view names to Things list names.
func ViewNameToListName(view string) (string, error) {
	mapping := map[string]string{
		"inbox":    "Inbox",
		"today":    "Today",
		"upcoming": "Upcoming",
		"anytime":  "Anytime",
		"someday":  "Someday",
		"logbook":  "Logbook",
		"trash":    "Trash",
		"tomorrow": "Tomorrow",
	}
	name, ok := mapping[strings.ToLower(view)]
	if !ok {
		return "", fmt.Errorf("unknown view: %s (valid: inbox, today, upcoming, anytime, someday, logbook, tomorrow)", view)
	}
	return name, nil
}

// ViewNameToShowID maps CLI view names to Things URL scheme show IDs.
func ViewNameToShowID(view string) (string, error) {
	mapping := map[string]string{
		"inbox":     "inbox",
		"today":     "today",
		"upcoming":  "upcoming",
		"anytime":   "anytime",
		"someday":   "someday",
		"logbook":   "logbook",
		"deadlines": "deadlines",
		"tomorrow":  "tomorrow",
	}
	id, ok := mapping[strings.ToLower(view)]
	if !ok {
		return "", fmt.Errorf("unknown view: %s (valid: inbox, today, upcoming, anytime, someday, logbook, deadlines, tomorrow)", view)
	}
	return id, nil
}
