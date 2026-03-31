package cmd

import (
	"fmt"
	"os"

	"github.com/danott/things-cli/internal/output"
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
	}

	cmd.AddCommand(newTagListCmd())
	cmd.AddCommand(newTagAddCmd())
	cmd.AddCommand(newTagRenameCmd())
	cmd.AddCommand(newTagDeleteCmd())

	return cmd
}

func newTagListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := getDB()
			if err != nil {
				return err
			}
			tags, err := db.ListTags()
			if err != nil {
				return err
			}

			if flagJSON {
				return output.PrintTagsJSON(os.Stdout, tags)
			}

			if len(tags) == 0 {
				fmt.Println("No tags found.")
				return nil
			}
			output.PrintTagsText(os.Stdout, tags)
			return nil
		},
	}
}
