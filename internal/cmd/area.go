package cmd

import (
	"fmt"
	"os"

	"github.com/danott/things-cli/internal/output"
	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newAreaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "area",
		Short: "Manage areas",
	}

	cmd.AddCommand(newAreaListCmd())
	cmd.AddCommand(newAreaAddCmd())
	cmd.AddCommand(newAreaRenameCmd())
	cmd.AddCommand(newAreaDeleteCmd())

	return cmd
}

func newAreaListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all areas",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := getDB()
			if err != nil {
				return err
			}
			areas, err := db.ListAreas()
			if err != nil {
				return err
			}

			if flagJSON {
				return output.PrintAreasJSON(os.Stdout, areas)
			}
			if flagMarkdown {
				output.PrintAreasMarkdown(os.Stdout, areas)
				return nil
			}

			if len(areas) == 0 {
				fmt.Println("No areas found.")
				return nil
			}
			output.PrintAreasText(os.Stdout, areas)
			return nil
		},
	}
}

func newAreaAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new area",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDryRun {
				fmt.Printf("AppleScript: make new area {name:%q}\n", args[0])
				return nil
			}
			return things.CreateArea(args[0])
		},
	}
}

func newAreaRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename an area",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDryRun {
				fmt.Printf("AppleScript: set name of area %q to %q\n", args[0], args[1])
				return nil
			}
			return things.RenameArea(args[0], args[1])
		},
	}
}

func newAreaDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an area",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDryRun {
				fmt.Printf("AppleScript: delete area %q\n", args[0])
				return nil
			}
			return things.DeleteArea(args[0])
		},
	}
}
