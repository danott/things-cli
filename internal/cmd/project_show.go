package cmd

import (
	"fmt"
	"os"

	"github.com/danott/things-cli/internal/output"
	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newProjectShowCmd() *cobra.Command {
	var flagGUI bool

	cmd := &cobra.Command{
		Use:   "show <title_or_id>",
		Short: "Show a project's details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if flagGUI {
				b := things.ShowItem(id)
				if flagDryRun {
					fmt.Println(b.Build())
					return nil
				}
				return b.Open()
			}

			db, err := getDB()
			if err != nil {
				return err
			}
			project, err := db.GetProject(id)
			if err != nil {
				return err
			}

			if flagJSON {
				return output.PrintProjectJSON(os.Stdout, project)
			}
			if flagMarkdown {
				output.PrintProjectMarkdown(os.Stdout, project)
				return nil
			}
			output.PrintProjectText(os.Stdout, project)
			return nil
		},
	}

	cmd.Flags().BoolVar(&flagGUI, "gui", false, "open in Things.app")

	return cmd
}
