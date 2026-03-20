package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

var (
	// version, commit and date are overwritten at build time.
	version string = "dev"
	commit  string = "unknown"
	date    string = "unknown"

	progName   string = "flight"
	flightRoot string = "/opt/flight"
	toolDir    string
)

func init() {
	log.SetReportTimestamp(false)
	log.SetReportCaller(false)
	log.SetLevel(log.WarnLevel)
	if root, ok := os.LookupEnv("FLIGHT_ROOT"); ok {
		flightRoot = root
	}
	toolDir = filepath.Join(flightRoot, "usr", "lib", "flight-core")
}

func main() {
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
				Name:     "tools",
				Usage:    "Manage Flight User Suite tools",
				Category: "Tool management",
				Commands: []*cli.Command{
					{
						Name:  "list",
						Usage: "List Flight User Suite tools",
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
						Arguments: []cli.Argument{
							&cli.StringArg{Name: "tool"},
						},
						ArgsUsage: "<tool>",
						Before:    assertArgPresent("tool"),
						Action:    enableTool,
					},
					{
						Name:  "disable",
						Usage: "Disable a Flight User Suite tool",
						Arguments: []cli.Argument{
							&cli.StringArg{Name: "tool"},
						},
						ArgsUsage: "<tool>",
						Before:    assertArgPresent("tool"),
						Action:    disableTool,
					},
				},
			},
		},
	}
	addToolProxyCommands(cmd)

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

// Assert that the named argument is present.
func assertArgPresent(argName string) cli.BeforeFunc {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		if cmd.NArg() == 0 {
			return ctx, MissingArguments{Args: []string{argName}}
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
		return fmt.Sprintf("Incorrect Usage: missing arguments %+v", ma.Args)
	}
}
