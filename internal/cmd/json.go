package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newJSONCmd() *cobra.Command {
	var (
		flagFile      string
		flagData      string
		flagReveal    bool
		flagAuthToken string
	)

	cmd := &cobra.Command{
		Use:   "json",
		Short: "Execute a Things JSON command",
		Long: `Send a JSON array of operations to Things via the json URL command.

Accepts JSON from --file, --data, or stdin. The JSON should be an array of
operation objects as documented at:
https://culturedcode.com/things/support/articles/2803573/

Auth token is only required if the JSON contains update operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var data string

			switch {
			case flagFile != "":
				b, readErr := os.ReadFile(flagFile)
				if readErr != nil {
					return fmt.Errorf("read file: %w", readErr)
				}
				data = string(b)
			case flagData != "":
				data = flagData
			default:
				// Read from stdin
				stat, _ := os.Stdin.Stat()
				if stat.Mode()&os.ModeCharDevice != 0 {
					return fmt.Errorf("provide JSON via --file, --data, or stdin")
				}
				b, readErr := io.ReadAll(os.Stdin)
				if readErr != nil {
					return fmt.Errorf("read stdin: %w", readErr)
				}
				data = string(b)
			}

			// Validate it's valid JSON
			if !json.Valid([]byte(data)) {
				return fmt.Errorf("invalid JSON")
			}

			// Resolve auth token (may be empty if only creating)
			authToken, _ := things.ResolveAuthToken(flagAuthToken)

			b := things.JSONCommand(data, authToken, flagReveal)
			if flagDryRun {
				fmt.Println(b.Build())
				return nil
			}
			return b.Open()
		},
	}

	cmd.Flags().StringVar(&flagFile, "file", "", "read JSON from file")
	cmd.Flags().StringVar(&flagData, "data", "", "inline JSON data")
	cmd.Flags().BoolVar(&flagReveal, "reveal", false, "navigate to first created item")
	cmd.Flags().StringVar(&flagAuthToken, "auth-token", "", "auth token (required for update operations)")

	return cmd
}
