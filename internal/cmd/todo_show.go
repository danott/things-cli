package cmd

import (
	"fmt"
	"os"

	"github.com/danott/things-cli/internal/config"
	"github.com/danott/things-cli/internal/interactive"
	"github.com/danott/things-cli/internal/output"
	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newTodoShowCmd() *cobra.Command {
	var flagGUI bool
	var flagInteractive bool

	cmd := &cobra.Command{
		Use:   "show <title_or_id>",
		Short: "Show a todo's details",
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

			if flagInteractive {
				// Resolve the todo ID (could be a title)
				todo, err := db.GetTodo(id)
				if err != nil {
					return err
				}
				authToken, err := things.ResolveAuthToken("")
				if err != nil {
					return err
				}
				cfg, err := config.Load()
				if err != nil {
					return err
				}
				return interactive.RunSingle(db, todo.ID, authToken, cfg.Actions)
			}

			todo, err := db.GetTodo(id)
			if err != nil {
				return err
			}

			if flagJSON {
				return output.PrintTodoJSON(os.Stdout, todo)
			}
			if flagMarkdown {
				output.PrintTodoMarkdown(os.Stdout, todo)
				return nil
			}
			output.PrintTodoText(os.Stdout, todo)
			return nil
		},
	}

	cmd.Flags().BoolVar(&flagGUI, "gui", false, "open in Things.app")
	cmd.Flags().BoolVarP(&flagInteractive, "interactive", "i", false, "interactive mode")

	return cmd
}
