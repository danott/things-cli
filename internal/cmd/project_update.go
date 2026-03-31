package cmd

import (
	"fmt"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newProjectUpdateCmd() *cobra.Command {
	var (
		opts      things.UpdateProjectOptions
		authToken string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing project",
		Long:  "Update a project by ID. Requires an auth token.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := things.ResolveAuthToken(authToken)
			if err != nil {
				return err
			}

			b := things.UpdateProject(args[0], token, opts)
			if flagDryRun {
				fmt.Println(b.Build())
				return nil
			}
			return b.Open()
		},
	}

	cmd.Flags().StringVar(&authToken, "auth-token", "", "Things auth token (overrides env/keychain)")
	cmd.Flags().StringVar(&opts.Title, "title", "", "new title")
	cmd.Flags().StringVar(&opts.Notes, "notes", "", "replace notes")
	cmd.Flags().StringVar(&opts.AppendNotes, "append-notes", "", "append to notes")
	cmd.Flags().StringVar(&opts.PrependNotes, "prepend-notes", "", "prepend to notes")
	cmd.Flags().StringVar(&opts.When, "when", "", "when to start")
	cmd.Flags().StringVar(&opts.Deadline, "deadline", "", "deadline (YYYY-MM-DD)")
	cmd.Flags().StringVar(&opts.Tags, "tags", "", "replace all tags (comma-separated)")
	cmd.Flags().StringVar(&opts.AddTags, "add-tags", "", "add tags without removing existing (comma-separated)")
	cmd.Flags().StringVar(&opts.Area, "area", "", "move to area by name")
	cmd.Flags().StringVar(&opts.AreaID, "area-id", "", "move to area by ID")
	cmd.Flags().BoolVar(&opts.Reveal, "reveal", false, "navigate to the updated project in Things")

	return cmd
}
