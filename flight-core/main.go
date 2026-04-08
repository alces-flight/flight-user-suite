package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strings"

	"charm.land/log/v2"
	"github.com/concertim/flight-user-suite/flight/pkg"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

var (
	// version, commit and date are overwritten at build time.
	version string = "dev"
	commit  string = "unknown"
	date    string = "unknown"

	progName        string = "flight"
	flightRoot      string = "/opt/flight"
	toolDir         string
	hookDir         string
	validEvents            = []string{"login", "activation"}
	validEventNames string = strings.Join(validEvents, ", ")
	termWidth       int    = 80
	maxTextWidth    int    = 80
)

func init() {
	log.SetReportTimestamp(false)
	log.SetReportCaller(false)
	log.SetLevel(log.WarnLevel)
	if root, ok := os.LookupEnv("FLIGHT_ROOT"); ok {
		flightRoot = root
	}
	var err error
	termWidth, _, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = 80
	}
	maxTextWidth = min(termWidth, 80)
	toolDir = filepath.Join(flightRoot, "usr", "lib", "flight-core")
	hookDir = filepath.Join(flightRoot, "usr", "lib", "hooks")
}

func main() {
	user, err := user.Current()
	if err != nil {
		log.Warn("Unable to determine user: not adding admin commands", "err", err)
	}

	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("version=%s revision=%s date=%s\n", version, commit, date)
	}
	cmd := &cli.Command{
		Name:                   progName,
		Usage:                  "The Flight User Suite",
		Version:                version,
		Suggest:                true,
		Copyright:              "(c) 2026 Stephen F Norledge & Concertim Ltd.",
		UseShortOptionHandling: true,
		HideHelpCommand:        true,
		Description:            rootDescription(user, maxTextWidth),
		CommandNotFound: func(ctx context.Context, cmd *cli.Command, command string) {
			fmt.Fprintf(
				cmd.Root().Writer,
				"%s: '%s' is not a %s command.  See '%s --help'.\n",
				progName, command, progName, progName,
			)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "set the log `LEVEL` (debug, info, warn, error, fatal). Default: warn",
				Validator: func(val string) error {
					switch strings.ToLower(val) {
					case "debug", "info", "warn", "error", "fatal":
					default:
						return fmt.Errorf("%s is not a known log level (debug, info, warn, error, fatal)", val)
					}
					return nil
				},
				Action: func(ctx context.Context, cmd *cli.Command, s string) error {
					level, err := log.ParseLevel(cmd.String("log-level"))
					if err != nil {
						panic("invalid value despite prior validation")
					}
					log.SetLevel(level)
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:        "shell",
				Usage:       "Enter a shell-like sandbox for running Flight tools",
				Description: wordwrap.String("Enter a shell-like sandbox for the 'flight' tool.", maxTextWidth),
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "tool", UsageText: "[TOOL]"},
				},
				Action: runShell,
			},
			configCommand(),
		},
	}
	if user != nil && user.Uid == "0" {
		addAdminCommands(cmd, maxTextWidth)
	}
	addToolProxyCommands(cmd)

	// Override help printer to inject some colour.
	origHelpPrinter := cli.HelpPrinter
	cli.HelpPrinter = pkg.ColourisedHelpPrinter(origHelpPrinter)

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		// A bunch of checks to avoid reporting the usage errors twice.
		errStr := err.Error()
		if (strings.Contains(errStr, "invalid value") && strings.Contains(errStr, "for flag")) ||
			(strings.Contains(errStr, "flag provided but not defined")) ||
			(strings.Contains(errStr, "flag needs an argument")) {
			// We've already reported the usage error.  No need to do so a
			// second time.
			os.Exit(1)
		}

		if strings.Contains(errStr, "cannot be set along with") {
			log.Printf("\nIncorrect Usage: %s", err)
			os.Exit(1)
		}

		if exitError, ok := errors.AsType[SilentExitError](err); ok {
			os.Exit(exitError.ExitCode)
		} else {
			log.Printf("%s\n", err)
			os.Exit(1)
		}
	}
}

func rootDescription(user *user.User, maxTextWidth int) string {
	var desc string
	if user != nil && user.Uid == "0" {
		desc = fmt.Sprintf(
			"Manage the Flight User Suite tools and hooks and access enabled tools.\n\nTools can be managed with the '%s tools' command and any enabled tools can be accessed as '%s <tool>'. A list of enabled tools can be found with '%s tools list --enabled'.\n\nHooks can be managed with the '%s hooks' command. Enabled hooks are executed either when a login shell is started or the Flight environment is activated.  See '%s hooks --help' for more details.",
			progName, progName, progName, progName, progName,
		)
	} else {
		desc = "Access Flight User Suite tools"
	}
	return wordwrap.String(desc, maxTextWidth)
}

func addAdminCommands(cmd *cli.Command, maxTextWidth int) {
	cmds := []*cli.Command{
		{
			Name:      "tools",
			Usage:     "Manage Flight User Suite tools",
			UsageText: fmt.Sprintf("%s tools command [command options]", progName),
			Description: wordwrap.String(
				fmt.Sprintf(
					"Manage the availability of the Flight User Suite tools.\n\nWhen a tool is enabled, it is available as '%s <tool>' and any howto guides it provides are made available through '%s howto'.",
					progName, progName,
				),
				maxTextWidth,
			),
			Category: "Tool management",
			Commands: []*cli.Command{
				{
					Name:        "list",
					Usage:       "List Flight User Suite tools",
					UsageText:   fmt.Sprintf("%s tools list [--enabled]", progName),
					Description: wordwrap.String("List all Flight User Suite tools.  If the --enabled flag is set, only list those tools that are currently enabled.", maxTextWidth),
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "enabled",
							Usage: "list only enabled tools",
						},
					},
					Action: listTools,
				},
				{
					Name:  "enable",
					Usage: "Enable a Flight User Suite tool",
					Description: wordwrap.String(
						fmt.Sprintf(
							"Enable the specified tool. When a tool is enabled, it is available as '%s <tool>' and any howto guides it provides are made available through '%s howto'.\n\nA list of available tools can be found with '%s tools list'.",
							progName, progName, progName,
						),
						maxTextWidth,
					),
					Arguments: []cli.Argument{
						&cli.StringArg{Name: "tool", UsageText: "<tool>"},
					},
					Before: assertArgPresent("tool"),
					Action: enableTool,
				},
				{
					Name:  "disable",
					Usage: "Disable a Flight User Suite tool",
					Description: wordwrap.String(
						fmt.Sprintf(
							"When a tool is disabled, it is no longer available as '%s <tool>' and any howto guides it provides are no longer available through '%s howto'.\n\nA list of currently enabled tools can be found with '%s tools list --enabled'.",
							progName, progName, progName,
						),
						maxTextWidth,
					),
					Arguments: []cli.Argument{
						&cli.StringArg{Name: "tool", UsageText: "<tool>"},
					},
					Before: assertArgPresent("tool"),
					Action: disableTool,
				},
			},
		},
		{
			Name:      "hooks",
			Usage:     "Manage Flight User Suite hooks",
			UsageText: fmt.Sprintf("%s hooks command [command options]", progName),
			Description: wordwrap.String(
				`Manage the Flight User Suite hooks.

There are two types of hooks: 'login' hooks and 'activation' hooks.  Enabled 'login' hooks are exectued when a login shell is started. Enabled 'activation' hooks are executed with the Flight environment is activated.`,
				maxTextWidth,
			),
			Category: "Hook management",
			Commands: []*cli.Command{
				{
					Name:  "list",
					Usage: "List Flight User Suite hooks",
					Description: wordwrap.String(
						fmt.Sprintf(
							`List all Flight User Suite hooks for the given event.  If the --enabled flag is set, only list those hooks that are currently enabled.  If the --event flag is provided, only list hooks for the given event.

Valid events are %s.`,
							validEventNames,
						),
						maxTextWidth,
					),
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "enabled",
							Usage: "list only enabled hooks",
						},
						&cli.StringFlag{
							Name:  "event",
							Usage: "list only hooks for the given event",
							Validator: func(event string) error {
								if event != "" && !slices.Contains(validEvents, event) {
									return fmt.Errorf("valid values are %s", validEventNames)
								}
								return nil
							},
						},
					},
					Action: listHooks,
				},
				{
					Name:  "enable",
					Usage: "Enable a Flight User Suite hook",
					Description: wordwrap.String(
						fmt.Sprintf(
							"Enable the specified <event> hook.  Valid events are %s.\n\nEnabled 'login' hooks are executed when a login shell is started.  Enabled 'activation' hooks are executed when the Flight environment is activated.\n\nA list of available hooks can be found with '%s hooks list <event>'.",
							validEventNames, progName,
						),
						maxTextWidth,
					),
					Arguments: []cli.Argument{
						&cli.StringArg{Name: "event", UsageText: "<event>"},
						&cli.StringArg{Name: "hook", UsageText: "<hook>"},
					},
					Before: composeBeforeFuncs(assertArgPresent("event", "hook"), assertEventValid("event", 0)),
					Action: enableHook,
				},
				{
					Name:  "disable",
					Usage: "Disable a Flight User Suite hook",
					Description: wordwrap.String(
						fmt.Sprintf(
							"Disable the specified <event> hook.  Valid events are %s.\n\nDisabled hooks are not executed when <event> occurs.\n\nA list of currently enabled hooks can be found with '%s hooks list --enabled <event>'.",
							validEventNames, progName,
						),
						maxTextWidth,
					),
					Arguments: []cli.Argument{
						&cli.StringArg{Name: "event", UsageText: "<event>"},
						&cli.StringArg{Name: "hook", UsageText: "<hook>"},
					},
					Before: composeBeforeFuncs(assertArgPresent("event", "hook"), assertEventValid("event", 0)),
					Action: disableHook,
				},
			},
		},
	}
	cmd.Commands = append(cmd.Commands, cmds...)
}

//go:embed tool_synopsis.txt
var toolSynopsisString string

func addToolProxyCommands(cmd *cli.Command) {
	synopsisMap := make(map[string]string)
	for line := range strings.Lines(toolSynopsisString) {
		parts := strings.SplitN(line, ":", 2)
		synopsisMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	tools, err := getTools(true)
	if err != nil {
		log.Warn("Unable to add tool proxy commands", "err", err)
		return
	}
	for _, tool := range tools {
		proxy := cli.Command{
			Name:            tool,
			Action:          runTool(tool),
			SkipFlagParsing: true,
			Category:        "Available tools",
		}
		if synopsis, found := synopsisMap[tool]; found {
			proxy.Usage = synopsis
		}
		cmd.Commands = append(cmd.Commands, &proxy)
	}
}

func composeBeforeFuncs(fns ...cli.BeforeFunc) cli.BeforeFunc {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		var err error
		for _, fn := range fns {
			ctx, err = fn(ctx, cmd)
			if err != nil {
				return ctx, err
			}
		}
		return ctx, nil
	}
}

// Assert that the named argument is present.
func assertArgPresent(argNames ...string) cli.BeforeFunc {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		if cmd.NArg() < len(argNames) {
			missing := argNames[cmd.NArg():]
			return ctx, MissingArguments{Args: missing}
		}
		return ctx, nil
	}
}

func assertEventValid(argName string, argIndex int) cli.BeforeFunc {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		event := cmd.Args().Get(argIndex)
		if !slices.Contains(validEvents, event) {
			return ctx, fmt.Errorf(
				"Incorrect Usage: unknown %s '%s'. Valid values are %s.",
				argName,
				event,
				validEventNames,
			)
		}
		return ctx, nil
	}
}

type MissingArguments struct {
	Args []string
}

func (ma MissingArguments) Error() string {
	if len(ma.Args) == 1 {
		return fmt.Sprintf("Incorrect Usage: missing argument %s", ma.Args[0])
	} else {
		return fmt.Sprintf("Incorrect Usage: missing arguments %s", strings.Join(ma.Args, ", "))
	}
}
