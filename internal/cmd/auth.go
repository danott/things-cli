package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage auth token",
		Long:  "Show auth token status or store the token in macOS Keychain.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(things.AuthStatus())
			return nil
		},
	}

	cmd.AddCommand(newAuthSetCmd())

	return cmd
}

func newAuthSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Store auth token in macOS Keychain",
		Long: `Store the Things auth token in macOS Keychain.

Find your token in Things > Settings > General > Enable Things URLs > Manage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("Enter your Things auth token: ")
			reader := bufio.NewReader(os.Stdin)
			token, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("read token: %w", err)
			}
			token = strings.TrimSpace(token)
			if token == "" {
				return fmt.Errorf("token cannot be empty")
			}

			if err := things.KeychainSet(token); err != nil {
				return err
			}
			fmt.Println("Auth token stored in macOS Keychain.")
			return nil
		},
	}
}
