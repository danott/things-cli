package cmd

import (
	"fmt"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newTagAddCmd() *cobra.Command {
	var parentTag string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDryRun {
				if parentTag != "" {
					fmt.Printf("AppleScript: make new tag {name:%q, parent tag:%q}\n", args[0], parentTag)
				} else {
					fmt.Printf("AppleScript: make new tag {name:%q}\n", args[0])
				}
				return nil
			}
			return things.CreateTag(args[0], parentTag)
		},
	}

	cmd.Flags().StringVar(&parentTag, "parent", "", "parent tag name")

	return cmd
}
