package cmd

import (
	"fmt"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newTagRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename a tag",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDryRun {
				fmt.Printf("AppleScript: set name of tag %q to %q\n", args[0], args[1])
				return nil
			}
			return things.RenameTag(args[0], args[1])
		},
	}
}
