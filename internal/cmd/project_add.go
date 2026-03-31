package cmd

import (
	"fmt"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newProjectAddCmd() *cobra.Command {
	var opts things.AddProjectOptions

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b := things.AddProject(args[0], opts)
			if flagDryRun {
				fmt.Println(b.Build())
				return nil
			}
			return b.Open()
		},
	}

	cmd.Flags().StringVar(&opts.Notes, "notes", "", "project notes")
	cmd.Flags().StringVar(&opts.When, "when", "", "when to start")
	cmd.Flags().StringVar(&opts.Deadline, "deadline", "", "deadline (YYYY-MM-DD)")
	cmd.Flags().StringVar(&opts.Tags, "tags", "", "comma-separated tag names")
	cmd.Flags().StringVar(&opts.Area, "area", "", "area name")
	cmd.Flags().StringVar(&opts.AreaID, "area-id", "", "area ID")
	cmd.Flags().StringVar(&opts.Todos, "todos", "", "newline-separated todo titles to create as children")
	cmd.Flags().BoolVar(&opts.Reveal, "reveal", false, "navigate to the created project in Things")

	return cmd
}
