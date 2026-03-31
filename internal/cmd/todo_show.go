package cmd

import (
	"fmt"
	"os"

	"github.com/danott/things-cli/internal/output"
	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newTodoShowCmd() *cobra.Command {
	var flagGUI bool

	cmd := &cobra.Command{
		Use:   "show <id>",
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
			todo, err := db.GetTodo(id)
			if err != nil {
				return err
			}

			if flagJSON {
				return output.PrintTodoJSON(os.Stdout, todo)
			}
			output.PrintTodoText(os.Stdout, todo)
			return nil
		},
	}

	cmd.Flags().BoolVar(&flagGUI, "gui", false, "open in Things.app")

	return cmd
}
