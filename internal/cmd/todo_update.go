package cmd

import (
	"fmt"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newTodoUpdateCmd() *cobra.Command {
	var (
		opts      things.UpdateTodoOptions
		authToken string
		flagEdit  bool
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing todo",
		Long:  "Update a todo by ID. Requires an auth token.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := things.ResolveAuthToken(authToken)
			if err != nil {
				return err
			}

			if flagEdit {
				db, err := getDB()
				if err != nil {
					return err
				}
				todo, err := db.GetTodo(args[0])
				if err != nil {
					return err
				}
				src, err := openInEditor(things.TodoToMarkdown(todo))
				if err != nil {
					return err
				}
				_, editedOpts, checklist, err := things.ParseTodoMarkdownForUpdate(src)
				if err != nil {
					return err
				}

				payload, err := things.BuildTodoUpdateJSON(args[0], editedOpts, checklist)
				if err != nil {
					return err
				}
				b := things.JSONCommand(payload, token, false)
				if flagDryRun {
					fmt.Println(b.Build())
					return nil
				}
				return b.Open()
			}

			b := things.UpdateTodo(args[0], token, opts)
			if flagDryRun {
				fmt.Println(b.Build())
				return nil
			}
			return b.Open()
		},
	}

	cmd.Flags().BoolVar(&flagEdit, "edit", false, "open $EDITOR with current todo contents")
	cmd.Flags().StringVar(&authToken, "auth-token", "", "Things auth token (overrides env/keychain)")
	cmd.Flags().StringVar(&opts.Title, "title", "", "new title")
	cmd.Flags().StringVar(&opts.Notes, "notes", "", "replace notes")
	cmd.Flags().StringVar(&opts.AppendNotes, "append-notes", "", "append to notes")
	cmd.Flags().StringVar(&opts.PrependNotes, "prepend-notes", "", "prepend to notes")
	cmd.Flags().StringVar(&opts.When, "when", "", "when to do it")
	cmd.Flags().StringVar(&opts.Deadline, "deadline", "", "deadline (YYYY-MM-DD)")
	cmd.Flags().StringVar(&opts.Tags, "tags", "", "replace all tags (comma-separated)")
	cmd.Flags().StringVar(&opts.AddTags, "add-tags", "", "add tags without removing existing (comma-separated)")
	cmd.Flags().StringVar(&opts.ChecklistItems, "checklist-items", "", "replace checklist items (newline-separated)")
	cmd.Flags().StringVar(&opts.AppendChecklistItems, "append-checklist-items", "", "append checklist items")
	cmd.Flags().StringVar(&opts.PrependChecklistItems, "prepend-checklist-items", "", "prepend checklist items")
	cmd.Flags().StringVar(&opts.List, "list", "", "move to project or area by name")
	cmd.Flags().StringVar(&opts.ListID, "list-id", "", "move to project or area by ID")
	cmd.Flags().StringVar(&opts.Heading, "heading", "", "move to heading by name")
	cmd.Flags().StringVar(&opts.HeadingID, "heading-id", "", "move to heading by ID")
	cmd.Flags().BoolVar(&opts.Reveal, "reveal", false, "navigate to the updated todo in Things")

	return cmd
}
