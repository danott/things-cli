package cmd

import (
	"fmt"

	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

func newVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("things-cli %s\n", version)

			thingsVersion, err := things.GetThingsVersion()
			if err != nil {
				fmt.Println("Things 3: not running or not installed")
			} else {
				fmt.Printf("Things 3:  %s\n", thingsVersion)
			}
			return nil
		},
	}
}
