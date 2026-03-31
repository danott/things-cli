package cmd

import (
	"fmt"
	"os"

	"github.com/danott/things-cli/internal/output"
	"github.com/spf13/cobra"
)

func newProjectListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := getDB()
			if err != nil {
				return err
			}
			projects, err := db.ListProjects()
			if err != nil {
				return err
			}

			if flagJSON {
				return output.PrintProjectsJSON(os.Stdout, projects)
			}

			if len(projects) == 0 {
				fmt.Println("No projects found.")
				return nil
			}
			output.PrintProjectsText(os.Stdout, projects)
			return nil
		},
	}
}
