package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"charm.land/log/v2"
	"github.com/cyucelen/marker"
	"github.com/fatih/color"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

var (
	// version, commit and date are overwritten at build time.
	version string = "dev"
	commit  string = "unknown"
	date    string = "unknown"

	progName     string = "flight desktop"
	flightRoot   string = "/opt/flight"
	termWidth    int    = 80
	maxTextWidth int    = 80
)

func init() {
	log.SetReportTimestamp(false)
	log.SetReportCaller(false)
	log.SetLevel(log.InfoLevel)
	if root, ok := os.LookupEnv("FLIGHT_ROOT"); ok {
		flightRoot = root
	}
	var err error
	termWidth, _, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = 80
	}
	maxTextWidth = min(termWidth, 80)
}

func main() {

	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Printf("version=%s revision=%s date=%s\n", version, commit, date)
	}
	cmd := &cli.Command{
		Name:                   progName,
		Usage:                  "Manage interactive GUI desktop sessions",
		Version:                version,
		Suggest:                true,
		Copyright:              "(c) 2026 Stephen F Norledge & Concertim Ltd.",
		UseShortOptionHandling: true,
		HideHelpCommand:        true,
		Description:            wordwrap.String("Manage interactive GUI desktop sessions", maxTextWidth),
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
			startCommand(),
			listSessionsCommand(),
			killSessionCommand(),
			showSessionCommand(),
			typeAvailCommand(),
		},
	}

	// Override help printer to inject some colour.
	origHelpPrinter := cli.HelpPrinter
	cli.HelpPrinter = func(w io.Writer, templ string, data any) {
		var buf bytes.Buffer
		origHelpPrinter(&buf, templ, data)
		bytes, err := io.ReadAll(&buf)
		if err != nil {
			log.Fatal("error formatting help output", "err", err)
		}
		headers := regexp.MustCompile("(?m:^[[:word:]].*:)")
		b := &marker.MarkBuilder{}
		ctmOrange := color.RGB(255, 116, 1)
		out := b.SetString(string(bytes)).
			Mark(marker.MatchRegexp(headers), ctmOrange).
			Build()
		fmt.Fprint(w, out)
	}

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
