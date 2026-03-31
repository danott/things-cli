package cmd

import "github.com/spf13/cobra"

func newTodoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "todo",
		Short: "Manage todos",
		Long:  "Create, list, show, update, complete, cancel, and delete todos.",
	}

	cmd.AddCommand(newTodoAddCmd())
	cmd.AddCommand(newTodoShowCmd())
	cmd.AddCommand(newTodoUpdateCmd())
	cmd.AddCommand(newTodoCompleteCmd())
	cmd.AddCommand(newTodoCancelCmd())
	cmd.AddCommand(newTodoIncompleteCmd())
	cmd.AddCommand(newTodoDeleteCmd())
	cmd.AddCommand(newTodoDuplicateCmd())

	return cmd
}
