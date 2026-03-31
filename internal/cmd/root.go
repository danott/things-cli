package cmd

import (
	"fmt"
	"os"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

var (
	flagDryRun   bool
	flagJSON     bool
	flagMarkdown bool
)

// Shared database connection, opened lazily on first read command.
var db *things.DB

func getDB() (*things.DB, error) {
	if db != nil {
		return db, nil
	}
	var err error
	db, err = things.OpenDB()
	return db, err
}

func NewRootCmd(version string) *cobra.Command {
	cobra.EnableCommandSorting = false

	cmd := &cobra.Command{
		Use:   "things",
		Short: "CLI for Things 3",
		Long:  "A command-line interface for Cultured Code's Things 3 task manager.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown command %q", args[0])
			}
			return cmd.Help()
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if db != nil {
				db.Close()
			}
		},
	}

	cmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "print the action without executing it")
	cmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "output as JSON")
	cmd.PersistentFlags().BoolVar(&flagMarkdown, "md", false, "output as Markdown")

	cmd.AddGroup(&cobra.Group{ID: "views", Title: "Views:"})
	cmd.AddGroup(&cobra.Group{ID: "manage", Title: "Commands:"})

	addCmd := func(c *cobra.Command, groupID string) {
		c.GroupID = groupID
		cmd.AddCommand(c)
	}

	for _, view := range views {
		addCmd(newViewCmd(view), "views")
	}
	addCmd(newTodoCmd(), "manage")
	addCmd(newProjectCmd(), "manage")
	addCmd(newAreaCmd(), "manage")
	addCmd(newTagCmd(), "manage")
	addCmd(newJSONCmd(), "manage")
	addCmd(newAuthCmd(), "manage")
	addCmd(newVersionCmd(version), "manage")

	return cmd
}

func Execute(version string) {
	cmd := NewRootCmd(version)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
