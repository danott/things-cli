package cmd

import (
	"fmt"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newTodoIncompleteCmd() *cobra.Command {
	var authToken string
	cmd := &cobra.Command{
		Use:   "incomplete <id>",
		Short: "Mark a completed or canceled todo as incomplete",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := things.ResolveAuthToken(authToken)
			if err != nil {
				return err
			}
			b := things.UpdateTodo(args[0], token, things.UpdateTodoOptions{Status: things.StatusPtr(things.StatusOpen)})
			if flagDryRun {
				fmt.Println(b.Build())
				return nil
			}
			return b.Open()
		},
	}
	cmd.Flags().StringVar(&authToken, "auth-token", "", "Things auth token (overrides env/keychain)")
	return cmd
}

func newTodoCompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <id>",
		Short: "Mark a todo as completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDryRun {
				fmt.Printf("AppleScript: set status of to do id \"%s\" to completed\n", args[0])
				return nil
			}
			return things.CompleteTodo(args[0])
		},
	}
}

func newTodoCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <id>",
		Short: "Mark a todo as canceled",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDryRun {
				fmt.Printf("AppleScript: set status of to do id \"%s\" to canceled\n", args[0])
				return nil
			}
			return things.CancelTodo(args[0])
		},
	}
}

func newTodoDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Move a todo to Trash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDryRun {
				fmt.Printf("AppleScript: delete to do id \"%s\"\n", args[0])
				return nil
			}
			return things.DeleteTodo(args[0])
		},
	}
}
