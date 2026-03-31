package main

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

var (
	// TODO: Determine this dynamically by listing the correct directory
	// (opt/flight/usr/lib/desktop/types/).
	validTypes            = []string{"terminal", "gnome"}
	validTypeNames string = strings.Join(validTypes, ", ")
)

func libexecPath(relpath string) string {
	return filepath.Join(flightRoot, "usr", "libexec", "desktop", relpath)
}

func startCommand() *cli.Command {
	return &cli.Command{
		Name:        "start",
		Usage:       "Start an interactive desktop session",
		Description: wordwrap.String("Start a new interactive desktop session and display details about the new session.\n\nAvailable desktop types can be shown using the 'avail' command.", maxTextWidth),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Name the desktop session `NAME` so it can be more easily identified.",
			},
			&cli.StringFlag{
				Name:    "geometry",
				Aliases: []string{"g"},
				Usage:   "Set the desktop geometry to `WIDTHxHEIGHT`.",
				Value:   "1024x768",
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "type", UsageText: "<type>"},
		},
		Before: composeBeforeFuncs(assertArgPresent("type"), assertTypeValid("type", 0)),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			sessionType := cmd.StringArg("type")
			fmt.Printf("Starting a '%s' desktop session:\n\n", sessionType)

			// TODO: Display a spinner.

			session := Session{
				UUID:         uuid.New(),
				SessionState: New,
				SessionType:  sessionType,
				Name:         cmd.String("name"),
				Geometry:     cmd.String("geometry"),
			}
			err := session.start(ctx)

			// TODO: Stop the spinner

			if err != nil {
				fmt.Printf("\u274c Starting session\n\n")
				return fmt.Errorf("starting session: %w", err)
			}
			fmt.Printf("\u2705 Starting session\n\n")
			fmt.Printf("A '%s' desktop session has been started.\n", session.SessionType)
			printSessionDetails(session)
			accessSummary(session)
			return nil
		},
	}
}

func assertTypeValid(argName string, argIndex int) cli.BeforeFunc {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {

		event := cmd.Args().Get(argIndex)
		if !slices.Contains(validTypes, event) {
			return ctx, fmt.Errorf(
				"Incorrect Usage: unknown %s '%s'. Valid values are %s.",
				argName,
				event,
				validTypeNames,
			)
		}
		return ctx, nil
	}
}
