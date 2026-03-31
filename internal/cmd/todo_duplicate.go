package cmd

import (
	"fmt"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newTodoDuplicateCmd() *cobra.Command {
	var (
		opts      things.UpdateTodoOptions
		authToken string
	)

	cmd := &cobra.Command{
		Use:   "duplicate <id>",
		Short: "Duplicate a todo",
		Long:  "Duplicate a todo by ID. Requires an auth token. Optional flags modify the copy.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := things.ResolveAuthToken(authToken)
			if err != nil {
				return err
			}
			opts.Duplicate = true
			b := things.UpdateTodo(args[0], token, opts)
			if flagDryRun {
				fmt.Println(b.Build())
				return nil
			}
			return b.Open()
		},
	}

	cmd.Flags().StringVar(&authToken, "auth-token", "", "Things auth token (overrides env/keychain)")
	cmd.Flags().StringVar(&opts.Title, "title", "", "title for the copy")
	cmd.Flags().StringVar(&opts.Notes, "notes", "", "notes for the copy")
	cmd.Flags().StringVar(&opts.When, "when", "", "when for the copy")
	cmd.Flags().StringVar(&opts.Deadline, "deadline", "", "deadline for the copy (YYYY-MM-DD)")
	cmd.Flags().StringVar(&opts.Tags, "tags", "", "tags for the copy (comma-separated)")
	cmd.Flags().StringVar(&opts.List, "list", "", "move copy to project or area by name")
	cmd.Flags().StringVar(&opts.ListID, "list-id", "", "move copy to project or area by ID")

	return cmd
}
