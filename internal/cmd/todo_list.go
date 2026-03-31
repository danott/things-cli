package cmd

import (
	"fmt"
	"os"

	"github.com/danott/things-cli/internal/output"
	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newTodoListCmd() *cobra.Command {
	var flagProject, flagArea, flagTag string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List todos filtered by project, area, or tag",
		RunE: func(cmd *cobra.Command, args []string) error {
			set := 0
			if flagProject != "" {
				set++
			}
			if flagArea != "" {
				set++
			}
			if flagTag != "" {
				set++
			}
			if set == 0 {
				return fmt.Errorf("one of --project, --area, or --tag is required")
			}
			if set > 1 {
				return fmt.Errorf("only one of --project, --area, or --tag may be specified")
			}

			db, err := getDB()
			if err != nil {
				return err
			}

			var todos []things.Todo
			switch {
			case flagProject != "":
				todos, err = db.ListProjectTodos(flagProject)
			case flagArea != "":
				todos, err = db.ListAreaTodos(flagArea)
			case flagTag != "":
				todos, err = db.ListTagTodos(flagTag)
			}
			if err != nil {
				return err
			}

			if flagJSON {
				return output.PrintTodosJSON(os.Stdout, todos)
			}
			if flagMarkdown {
				output.PrintTodosMarkdown(os.Stdout, todos)
				return nil
			}

			if len(todos) == 0 {
				fmt.Println("No todos found.")
				return nil
			}
			output.PrintTodosText(os.Stdout, todos, false)
			return nil
		},
	}

	cmd.Flags().StringVar(&flagProject, "project", "", "filter by project title or ID")
	cmd.Flags().StringVar(&flagArea, "area", "", "filter by area title or ID")
	cmd.Flags().StringVar(&flagTag, "tag", "", "filter by tag title or ID")

	return cmd
}
