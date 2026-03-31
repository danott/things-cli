package cmd

import (
	"fmt"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newTagDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDryRun {
				fmt.Printf("AppleScript: delete tag %q\n", args[0])
				return nil
			}
			return things.DeleteTag(args[0])
		},
	}
}
