package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/concertim/flight-user-suite/flight/pkg"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	flightRoot       string = "/opt/flight"
	howtoDir         string
	markdownThemeDir string
	termWidth        int = 80
	maxTextWidth     int = 80
)

func init() {
	if root, ok := os.LookupEnv("FLIGHT_ROOT"); ok {
		flightRoot = root
	}
	howtoDir = filepath.Join(flightRoot, "usr", "share", "doc", "howtos-enabled")
	markdownThemeDir = filepath.Join(flightRoot, "usr", "lib", "flight-howto", "themes")
	var err error
	termWidth, _, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = 80
	}
	maxTextWidth = min(termWidth, 80)
}

func main() {
	cmd := &cli.Command{
		Name:        "flight howto",
		Usage:       "View user guides for your HPC environment",
		Description: lipgloss.Wrap("View user guides for your HPC environment", maxTextWidth, " "),
		Copyright:   "(c) 2026 Stephen F Norledge & Concertim Ltd.",
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"l", "ls"},
				Usage:   "List available user guides",
				Action:  list,
			},
			{
				Name:      "show",
				Aliases:   []string{"s"},
				Usage:     "Open a user guide for viewing in the terminal",
				ArgsUsage: "<guide-name>",
				Action:    show,
				Before:    assertArgPresent("guide-name"),
			},
		},
	}

	// Override help printer to inject some colour.
	origHelpPrinter := cli.HelpPrinter
	cli.HelpPrinter = pkg.ColourisedHelpPrinter(origHelpPrinter)

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

	if !strings.HasSuffix(fullPath, ".md") {
		fullPath = fullPath + ".md"
	}

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
	markdownTheme := filepath.Join(markdownThemeDir, "flight-light.json")
	if isDark {
		markdownTheme = filepath.Join(markdownThemeDir, "flight-dark.json")
	}

	rendered, err := glamour.Render(string(markdown), markdownTheme)
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

	filenames := make([]string, len(entries))
	for i, entry := range entries {
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
				name, _ = strings.CutSuffix(relPath, ".md")
				filenames[i] = name
			}
		}
	}
	err = entriesTable(filenames)
	return err
}

func prettyFilename(filename string) (title string) {
	return cases.
		Title(language.English, cases.Compact).
		String(strings.ReplaceAll(filename, "-", " "))
}

func entriesTable(filenames []string) error {
	namecolWidth := 4
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(pkg.AlcesBlue)).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			switch {
			case row == table.HeaderRow:
				return pkg.TableHeaderStyle
			case row%2 == 0:
				style = pkg.TableEvenRowStyle
			default:
				style = pkg.TableOddRowStyle
			}
			switch col {
			case 0:
				return style.Width(namecolWidth)
			}
			return style
		}).
		Width(termWidth)
	t.Headers("ID", "Title")
	for index, name := range filenames {
		id := strconv.Itoa(index + 1)
		namecolWidth = max(namecolWidth, len(id)+2)
		titleColumn := lipgloss.JoinVertical(
			lipgloss.Left,
			prettyFilename(name),
		)
		t.Row(id, titleColumn)
	}
	_, err := lipgloss.Println(t)
	return err
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
