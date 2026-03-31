package cmd

import (
	"fmt"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newTodoAddCmd() *cobra.Command {
	var (
		opts    things.AddTodoOptions
		flagEdit bool
	)

	cmd := &cobra.Command{
		Use:   "add [title]",
		Short: "Add a new todo",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagEdit {
				initial := things.BlankTodoMarkdown()
				if len(args) == 1 {
					// Pre-populate title if provided
					initial = "---\nwhen: \ndeadline: \ntags: \nlist: \n---\n\n# " + args[0] + "\n"
				}
				src, err := openInEditor(initial)
				if err != nil {
					return err
				}
				title, editedOpts, err := things.ParseTodoMarkdown(src)
				if err != nil {
					return err
				}
				if title == "" {
					return fmt.Errorf("todo title is required")
				}
				b := things.AddTodo(title, editedOpts)
				if flagDryRun {
					fmt.Println(b.Build())
					return nil
				}
				return b.Open()
			}

			if len(args) == 0 {
				return fmt.Errorf("title required (or use --edit)")
			}
			b := things.AddTodo(args[0], opts)
			if flagDryRun {
				fmt.Println(b.Build())
				return nil
			}
			return b.Open()
		},
	}

	cmd.Flags().BoolVar(&flagEdit, "edit", false, "open $EDITOR to compose the todo")
	cmd.Flags().StringVar(&opts.Notes, "notes", "", "notes for the todo")
	cmd.Flags().StringVar(&opts.When, "when", "", "when to do it (today, tomorrow, evening, anytime, someday, or YYYY-MM-DD)")
	cmd.Flags().StringVar(&opts.Deadline, "deadline", "", "deadline (YYYY-MM-DD)")
	cmd.Flags().StringVar(&opts.Tags, "tags", "", "comma-separated tag names")
	cmd.Flags().StringVar(&opts.ChecklistItems, "checklist-items", "", "newline-separated checklist items")
	cmd.Flags().StringVar(&opts.List, "list", "", "project or area name to add to")
	cmd.Flags().StringVar(&opts.ListID, "list-id", "", "project or area ID to add to")
	cmd.Flags().StringVar(&opts.Heading, "heading", "", "heading within a project")
	cmd.Flags().StringVar(&opts.HeadingID, "heading-id", "", "heading ID within a project")
	cmd.Flags().BoolVar(&opts.Reveal, "reveal", false, "navigate to the created todo in Things")
	cmd.Flags().BoolVar(&opts.ShowQuickEntry, "show-quick-entry", false, "show Quick Entry dialog instead of adding directly")

	return cmd
}
