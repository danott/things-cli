package cmd

import "github.com/spf13/cobra"

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
	}

	cmd.AddCommand(newProjectAddCmd())
	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectShowCmd())
	cmd.AddCommand(newProjectUpdateCmd())
	cmd.AddCommand(newProjectCompleteCmd())
	cmd.AddCommand(newProjectCancelCmd())
	cmd.AddCommand(newProjectDeleteCmd())

	return cmd
}
