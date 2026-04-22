package main

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
	"github.com/yarlson/pin"
)

var (
	nameWhitelist            = "-_.A-Za-z0-9"
	nameWhitelistExplanation = "letters, numbers, hyphens, underscores and dots"
	nameBlacklist            = regexp.MustCompile(fmt.Sprintf("[^%s]+", nameWhitelist))
	nameMaxLen               = 40
)

func libexecPath(relpath string) string {
	return filepath.Join(env.FlightRoot, "usr", "libexec", "desktop", relpath)
}

func startSessionCommand() *cli.Command {
	return &cli.Command{
		Name:        "start",
		Usage:       "Start an interactive desktop session",
		Description: wordwrap.String("Start a new interactive desktop session and display details about the new session.\n\nAvailable desktop types can be shown using the 'avail' command.", maxTextWidth),
		Category:    "Sessions",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "name",
				Aliases:     []string{"n"},
				Usage:       "Name the desktop session `NAME` so it can be more easily identified.",
				DefaultText: "random",
				Validator: func(name string) error {
					if nameBlacklist.MatchString(name) {
						return fmt.Errorf("it can contain only %s.", nameWhitelistExplanation)
					}
					if len(name) > nameMaxLen {
						return fmt.Errorf("it must be no more than %d characters", nameMaxLen)
					}
					return nil
				},
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
		ShellComplete: func(ctx context.Context, cmd *cli.Command) {
			cli.DefaultCompleteWithFlags(ctx, cmd)
			switch cmd.NArg() {
			case 0:
				types, err := loadAllTypes()
				if err != nil {
					return
				}
				for _, t := range types {
					fmt.Println(t.ID)
				}
			}
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			sessionType := cmd.StringArg("type")
			name := cmd.String("name")
			if name == "" {
				name = newNameGenerator(sessionType).Generate()
			}
			fmt.Printf("Starting '%s' desktop session '%s':\n\n", sessionType, name)

			depsOK, err := checkDependencies(ctx, sessionType)

			if !depsOK {
				return err
			}

			p := createPin("Starting session...")
			cancel := p.Start(ctx)
			defer cancel()

			session := Session{
				Name:        name,
				State:       New,
				SessionType: sessionType,
				Geometry:    cmd.String("geometry"),
			}
			err = session.Start(ctx)
			if err != nil {
				p.Fail("Starting session failed")
				return err
			}
			p.Stop(fmt.Sprintf("Your %s session is ready!", session.SessionType))
			fmt.Println()
			sessionStarted(&session)
			connectionInfo(&session)
			managementInfo(&session)
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

func checkDependencies(ctx context.Context, sessionType string) (bool, error) {
	p := createPin("Checking system dependencies...")
	cancel := p.Start(ctx)
	defer cancel()

	// Add a small delay to stop the spinner from flickering
	<-time.After(1 * time.Second)

	globalResults, globalDepsOK := runDoctor(requiredDependencies(config.Dependencies))

	if !globalDepsOK {
		p.Fail("Missing critical dependencies")
		printCheckResults(globalResults)
		return false, nil
	}

	sessionTypeDef, err := loadType(sessionType)

	if err != nil {
		return false, err
	}

	err = sessionTypeDef.loadDependencies()
	if err != nil {
		return false, err
	}

	typeResults, typeDepsOK := runDoctor(requiredDependencies(sessionTypeDef.dependencies))

	if !typeDepsOK {
		p.Fail(fmt.Sprintf("Missing required dependencies for %s desktop type", sessionType))
		printCheckResults(typeResults)
		return false, err
	}

	p.Stop("Dependencies OK")

	return true, err
}

func createPin(text string) *pin.Pin {
	return pin.New(
		text,
		pin.WithSpinnerColor(pin.ColorCyan),
		pin.WithTextColor(pin.ColorGreen),
		pin.WithDoneSymbol('\u2705'),
		pin.WithFailSymbol('\u274c'),
		pin.WithFailColor(pin.ColorRed),
	)
}
