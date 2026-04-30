package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/danott/things-cli/internal/things"
)

func PrintTodosText(w io.Writer, todos []things.Todo, verbose bool) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	prevBucket := -1
	prevHeading := ""
	hasHeadings := false
	for _, t := range todos {
		if t.HeadingID != "" {
			hasHeadings = true
			break
		}
	}
	for _, t := range todos {
		if t.StartBucket == things.StartBucketEvening && prevBucket != things.StartBucketEvening {
			tw.Flush()
			fmt.Fprintf(w, "\nThis Evening\n\n")
			tw = tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			prevHeading = ""
		}
		prevBucket = t.StartBucket
		if hasHeadings && t.HeadingID != prevHeading {
			tw.Flush()
			if t.HeadingName != "" {
				fmt.Fprintf(w, "\n%s\n\n", t.HeadingName)
			} else {
				fmt.Fprintln(w)
			}
			tw = tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			prevHeading = t.HeadingID
		}
		check := "  "
		if t.Status == things.StatusCompleted {
			check = "x "
		} else if t.Status == things.StatusCanceled {
			check = "- "
		}
		if verbose {
			fmt.Fprintf(tw, "%s%s\t%s\n", check, t.Name, t.ID)
		} else {
			fmt.Fprintf(tw, "%s%s\n", check, t.Name)
		}
	}
	tw.Flush()
}

func PrintTodosMarkdown(w io.Writer, todos []things.Todo) {
	prevBucket := -1
	prevHeading := ""
	hasHeadings := false
	for _, t := range todos {
		if t.HeadingID != "" {
			hasHeadings = true
			break
		}
	}
	for _, t := range todos {
		if t.StartBucket == things.StartBucketEvening && prevBucket != things.StartBucketEvening {
			fmt.Fprintf(w, "\n## This Evening\n\n")
			prevHeading = ""
		}
		prevBucket = t.StartBucket
		if hasHeadings && t.HeadingID != prevHeading {
			if t.HeadingName != "" {
				fmt.Fprintf(w, "\n### %s\n\n", t.HeadingName)
			} else {
				fmt.Fprintln(w)
			}
			prevHeading = t.HeadingID
		}
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
	fmt.Fprintln(w, "---")
	fmt.Fprintf(w, "id: %s\n", t.ID)
	fmt.Fprintf(w, "status: %s\n", t.Status)
	if t.ProjectName != "" {
		fmt.Fprintf(w, "project: %s\n", t.ProjectName)
	}
	if t.ProjectID != "" {
		fmt.Fprintf(w, "project-id: %s\n", t.ProjectID)
	}
	if t.AreaName != "" {
		fmt.Fprintf(w, "area: %s\n", t.AreaName)
	}
	if t.AreaID != "" {
		fmt.Fprintf(w, "area-id: %s\n", t.AreaID)
	}
	if t.TagNames != "" {
		fmt.Fprintf(w, "tags: %s\n", t.TagNames)
	}
	if t.ActivationDate != nil {
		fmt.Fprintf(w, "when: %s\n", t.ActivationDate.Format("2006-01-02"))
	}
	if t.DueDate != nil {
		fmt.Fprintf(w, "deadline: %s\n", t.DueDate.Format("2006-01-02"))
	}
	if t.StartBucket == things.StartBucketEvening {
		fmt.Fprintln(w, "start-bucket: evening")
	} else if t.StartBucket == things.StartBucketMorning && t.ActivationDate != nil {
		fmt.Fprintln(w, "start-bucket: morning")
	}
	if t.CreationDate != nil {
		fmt.Fprintf(w, "created: %s\n", t.CreationDate.Format(time.RFC3339))
	}
	if t.ModificationDate != nil {
		fmt.Fprintf(w, "modified: %s\n", t.ModificationDate.Format(time.RFC3339))
	}
	if t.CompletionDate != nil {
		fmt.Fprintf(w, "completed: %s\n", t.CompletionDate.Format(time.RFC3339))
	}
	if t.CancellationDate != nil {
		fmt.Fprintf(w, "canceled: %s\n", t.CancellationDate.Format(time.RFC3339))
	}
	fmt.Fprintln(w, "---")
	fmt.Fprintf(w, "\n# %s\n", t.Name)
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
			if item.ID != "" {
				fmt.Fprintf(w, "- [%s] %s (%s)\n", itemCheck, item.Name, item.ID)
			} else {
				fmt.Fprintf(w, "- [%s] %s\n", itemCheck, item.Name)
			}
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
	fmt.Fprintln(w, "---")
	fmt.Fprintf(w, "id: %s\n", p.ID)
	fmt.Fprintf(w, "status: %s\n", p.Status)
	if p.AreaName != "" {
		fmt.Fprintf(w, "area: %s\n", p.AreaName)
	}
	if p.TagNames != "" {
		fmt.Fprintf(w, "tags: %s\n", p.TagNames)
	}
	if p.ActivationDate != nil {
		fmt.Fprintf(w, "when: %s\n", p.ActivationDate.Format("2006-01-02"))
	}
	if p.DueDate != nil {
		fmt.Fprintf(w, "deadline: %s\n", p.DueDate.Format("2006-01-02"))
	}
	fmt.Fprintln(w, "---")
	fmt.Fprintf(w, "\n# %s\n", p.Name)
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
