package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/cyucelen/marker"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

var (
	flightRoot string = "/opt/flight"
	howtoDir   string
	themeDir   string
)

func init() {
	if root, ok := os.LookupEnv("FLIGHT_ROOT"); ok {
		flightRoot = root
	}
	howtoDir = filepath.Join(flightRoot, "usr", "share", "doc", "howtos-enabled")
	themeDir = filepath.Join(flightRoot, "usr", "lib", "flight-howto", "themes")
}

func main() {
	cmd := &cli.Command{
		Name:  "flight howto",
		Usage: "List and display Flight User Suite documentation",
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"l", "ls"},
				Usage:   "list available howtos",
				Action:  list,
			},
			{
				Name:      "show",
				Aliases:   []string{"s"},
				Usage:     "show a howto",
				ArgsUsage: "<howto>",
				Action:    show,
				Before:    assertArgPresent("howto"),
			},
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

	// TODO deduplicate this from equivalent section in flight-core/main.go?
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
			log.Printf("\nIncorrect usage: %s", err)
			os.Exit(1)
		}

		log.Printf("%s\n", err)
		os.Exit(1)
	}
}

func list(ctx context.Context, cmd *cli.Command) error {
	return PrintDirContents(howtoDir)
}

func show(ctx context.Context, cmd *cli.Command) error {
	howtoName := cmd.Args().First()
	fullPath := filepath.Join(howtoDir, howtoName)
	markdown, err := os.ReadFile(fullPath)
	if err != nil {
		if pathError, ok := errors.AsType[*fs.PathError](err); ok {
			if pathError.Err.Error() == "no such file or directory" {
				return UnknownHowto{Howto: howtoName}
			}
		}
		return fmt.Errorf("reading howto: %w", err)
	}

	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	theme := filepath.Join(themeDir, "flight-light.json")
	if isDark {
		theme = filepath.Join(themeDir, "flight-dark.json")
	}

	rendered, err := glamour.Render(string(markdown), theme)
	if err != nil {
		return fmt.Errorf("rendering howto: %w", err)
	}

	fmt.Print(rendered)
	return nil
}

func PrintDirContents(dirPath string) error {
	entries, err := os.ReadDir(dirPath)

	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		filePath := fmt.Sprintf("%v/%v", dirPath, name)

		if entry.IsDir() {
			PrintDirContents(filePath)
		} else {
			relPath, err := filepath.Rel(howtoDir, filePath)

			if err != nil {
				return fmt.Errorf("reading directory: %w", err)
			}

			ext := filepath.Ext(relPath)
			if ext == ".md" {
				fmt.Println(relPath)
			}
		}
	}
	return nil
}

// TODO properly share these with flight-core
type MissingArguments struct {
	Args []string
}

func (ma MissingArguments) Error() string {
	if len(ma.Args) == 1 {
		return fmt.Sprintf("Incorrect usage: missing argument %s", ma.Args[0])
	} else {
		return fmt.Sprintf("Incorrect usage: missing arguments %s", strings.Join(ma.Args, ", "))
	}
}

func assertArgPresent(argNames ...string) cli.BeforeFunc {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		if cmd.NArg() < len(argNames) {
			missing := argNames[cmd.NArg():]
			return ctx, MissingArguments{Args: missing}
		}
		return ctx, nil
	}
}

type UnknownHowto struct {
	Howto string
}

func (ut UnknownHowto) Error() string {
	return fmt.Sprintf("Unknown howto: %s", ut.Howto)
}
