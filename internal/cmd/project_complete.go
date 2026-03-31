package cmd

import (
	"fmt"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newProjectCompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <id>",
		Short: "Mark a project as completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDryRun {
				fmt.Printf("AppleScript: set status of project id \"%s\" to completed\n", args[0])
				return nil
			}
			return things.CompleteProject(args[0])
		},
	}
}

func newProjectCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <id>",
		Short: "Mark a project as canceled",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDryRun {
				fmt.Printf("AppleScript: set status of project id \"%s\" to canceled\n", args[0])
				return nil
			}
			return things.CancelProject(args[0])
		},
	}
}

func newProjectDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Move a project to Trash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDryRun {
				fmt.Printf("AppleScript: delete project id \"%s\"\n", args[0])
				return nil
			}
			return things.DeleteProject(args[0])
		},
	}
}
