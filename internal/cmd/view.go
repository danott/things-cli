package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/danott/things-cli/internal/config"
	"github.com/danott/things-cli/internal/interactive"
	"github.com/danott/things-cli/internal/output"
	"github.com/danott/things-cli/internal/things"
	"github.com/spf13/cobra"
)

var views = []string{"inbox", "today", "upcoming", "anytime", "someday", "logbook", "trash"}

func newViewCmd(name string) *cobra.Command {
	var flagGUI bool
	var flagInteractive bool
	var flagVerbose bool

	// View-specific flags
	var flagMorning, flagEvening bool
	var flagLimit, flagOffset int
	var flagSince, flagUntil string

	cmd := &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("Show todos in %s", name),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagInteractive {
				db, err := getDB()
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
				return interactive.Run(db, name, authToken, cfg.Actions)
			}

			if flagGUI {
				showID, err := output.ViewNameToShowID(name)
				if err != nil {
					return err
				}
				b := things.ShowList(showID)
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

			var todos []things.Todo

			switch {
			case name == "logbook":
				opts, parseErr := buildLogbookOptions(flagLimit, flagOffset, flagSince, flagUntil)
				if parseErr != nil {
					return parseErr
				}
				todos, err = db.ListLogbook(opts)
			case name == "today" && flagMorning:
				todos, err = db.ListTodos("today-morning")
			case name == "today" && flagEvening:
				todos, err = db.ListTodos("today-evening")
			default:
				todos, err = db.ListTodos(name)
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
				fmt.Printf("No todos in %s.\n", name)
				return nil
			}
			output.PrintTodosText(os.Stdout, todos, flagVerbose)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&flagInteractive, "interactive", "i", false, "interactive mode")
	cmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "show todo IDs")
	cmd.Flags().BoolVar(&flagGUI, "gui", false, "navigate to the view in Things.app")

	if name == "today" {
		cmd.Flags().BoolVar(&flagMorning, "morning", false, "show morning items only")
		cmd.Flags().BoolVar(&flagEvening, "evening", false, "show evening items only")
	}

	if name == "logbook" {
		cmd.Flags().IntVar(&flagLimit, "limit", 0, "max results (default: 50)")
		cmd.Flags().IntVar(&flagOffset, "offset", 0, "skip first N results")
		cmd.Flags().StringVar(&flagSince, "since", "", "filter by completion date >= YYYY-MM-DD")
		cmd.Flags().StringVar(&flagUntil, "until", "", "filter by completion date <= YYYY-MM-DD")
	}

	return cmd
}

func buildLogbookOptions(limit, offset int, since, until string) (things.LogbookOptions, error) {
	opts := things.LogbookOptions{
		Limit:  limit,
		Offset: offset,
	}
	if since != "" {
		t, err := time.ParseInLocation("2006-01-02", since, time.Local)
		if err != nil {
			return opts, fmt.Errorf("--since: invalid date %q (want YYYY-MM-DD)", since)
		}
		opts.Since = &t
	}
	if until != "" {
		t, err := time.ParseInLocation("2006-01-02", until, time.Local)
		if err != nil {
			return opts, fmt.Errorf("--until: invalid date %q (want YYYY-MM-DD)", until)
		}
		t = t.Add(24*time.Hour - time.Second)
		opts.Until = &t
	}
	return opts, nil
}
