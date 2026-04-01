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
	"github.com/yarlson/pin"
)

func libexecPath(relpath string) string {
	return filepath.Join(flightRoot, "usr", "libexec", "desktop", relpath)
}

func startSessionCommand() *cli.Command {
	return &cli.Command{
		Name:        "start",
		Usage:       "Start an interactive desktop session",
		Description: wordwrap.String("Start a new interactive desktop session and display details about the new session.\n\nAvailable desktop types can be shown using the 'avail' command.", maxTextWidth),
		Category:    "Sessions",
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

			p := pin.New("Starting session...",
				pin.WithSpinnerColor(pin.ColorCyan),
				pin.WithTextColor(pin.ColorGreen),
				pin.WithDoneSymbol('\u2705'),
				pin.WithFailSymbol('\u274c'),
				pin.WithFailColor(pin.ColorRed),
			)
			cancel := p.Start(ctx)
			defer cancel()

			session := Session{
				ID:           uuid.New().String(),
				SessionState: New,
				SessionType:  sessionType,
				Name:         cmd.String("name"),
				Geometry:     cmd.String("geometry"),
			}
			err := session.Start(ctx)
			if err != nil {
				p.Fail("Starting session failed")
				return fmt.Errorf("starting session: %w", err)
			}
			p.Stop(fmt.Sprintf("Your %s session is ready!", session.SessionType))
			fmt.Println()
			sessionStarted(&session)
			connectionInfo(&session)
			return nil
		},
	}
}

func assertTypeValid(argName string, argIndex int) cli.BeforeFunc {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		availableTypes, err := loadAllTypes()
		typeNames := make([]string, 0, len(availableTypes))
		for _, typ := range availableTypes {
			typeNames = append(typeNames, typ.ID)
		}
		if err != nil {
			return ctx, err
		}
		typ := cmd.Args().Get(argIndex)
		if !slices.Contains(typeNames, typ) {
			return ctx, fmt.Errorf(
				"Incorrect Usage: unknown %s '%s'. Valid values are %s.",
				argName,
				typ,
				strings.Join(typeNames, ", "),
			)
		}
		return ctx, nil
	}
}
