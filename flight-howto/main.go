package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"charm.land/log/v2"
	"github.com/concertim/flight-user-suite/flight/cliui"
	"github.com/concertim/flight-user-suite/flight/configenv"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

var (
	env              configenv.Env
	howtoDir         string
	markdownThemeDir string
	termWidth        int = 80
	maxTextWidth     int = 80
)

var formatFlag = &cli.StringFlag{
	Name:  "format",
	Value: "pretty",
	Usage: "use specified `FORMAT` for the output (pretty, json).",
	Validator: func(format string) error {
		if format != "pretty" && format != "json" {
			return fmt.Errorf("%s is not a known format (pretty, json)", format)
		}
		return nil
	},
}

func init() {
	log.SetReportTimestamp(false)
	log.SetReportCaller(false)
	log.SetLevel(log.WarnLevel)
	var err error
	env, err = configenv.InitFlightEnv()
	if err != nil {
		panic(fmt.Errorf("initializing flight env: %w", err))
	}
	howtoDir = filepath.Join(env.FlightRoot, "usr", "share", "doc", "howtos-enabled")
	markdownThemeDir = filepath.Join(env.FlightRoot, "usr", "lib", "flight-howto", "themes")
	termWidth, _, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = 80
	}
	maxTextWidth = min(termWidth, 80)
}

func main() {
	cmd := &cli.Command{
		Name:                  "flight howto",
		Usage:                 "View user guides for your HPC environment",
		Description:           lipgloss.Wrap("View user guides for your HPC environment", maxTextWidth, " "),
		Copyright:             "(c) 2026 Stephen F Norledge & Alces Software Ltd & Concertim Ltd.",
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"l", "ls"},
				Usage:   "List available user guides",
				Flags:   []cli.Flag{formatFlag},
				Action:  list,
			},
			{
				Name:      "show",
				Aliases:   []string{"s"},
				Usage:     "Open a user guide for viewing in the terminal",
				ArgsUsage: "<index>",
				Flags:     []cli.Flag{formatFlag},
				Action:    show,
				Before:    assertArgPresent("index"),
			},
		},
	}

	// Override help printer to inject some colour.
	origHelpPrinter := cli.HelpPrinter
	cli.HelpPrinter = cliui.ColourisedHelpPrinter(origHelpPrinter)

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

		if exitError, ok := errors.AsType[SilentExitError](err); ok {
			os.Exit(exitError.ExitCode)
		}

		log.Printf("%s\n", err)
		os.Exit(1)
	}
}

func list(ctx context.Context, cmd *cli.Command) error {
	wantsJSON := wantsJSONOutput(cmd)
	user, err := user.Current()
	if err != nil {
		log.Warn("Unable to determine user: including admin guides", "err", err)
	}
	howtos, err := loadHowtos(howtoDir, user)
	if err != nil {
		if wantsJSON {
			return writeListHowtosJSONError(err)
		}
		return err
	}
	if wantsJSON {
		return writeListHowtosJSON(howtos)
	}
	return entriesTable(howtos)
}

func show(ctx context.Context, cmd *cli.Command) error {
	wantsJSON := wantsJSONOutput(cmd)
	user, err := user.Current()
	if err != nil {
		log.Warn("Unable to determine user: including admin guides", "err", err)
	}
	howtos, err := loadHowtos(howtoDir, user)
	if err != nil {
		err = fmt.Errorf("collecting guide files: %w", err)
		if wantsJSON {
			return writeShowHowtoJSON(nil, err)
		}
		return err
	}

	howtoIndex, err := strconv.Atoi(cmd.Args().First())
	if err != nil || howtoIndex < 1 || howtoIndex > len(howtos) {
		err = InvalidIndex{Input: cmd.Args().First()}
		if wantsJSON {
			return writeShowHowtoJSON(nil, err)
		}
		return err
	}

	howto := howtos[howtoIndex-1]
	markdown, err := howto.Content()
	if err != nil {
		if pathError, ok := errors.AsType[*fs.PathError](err); ok {
			if pathError.Err.Error() == "no such file or directory" {
				err = UnknownHowto{Howto: howto.Path}
				if wantsJSON {
					return writeShowHowtoJSON(nil, err)
				}
				return err
			}
		}
		err = fmt.Errorf("reading howto: %w", err)
		if wantsJSON {
			return writeShowHowtoJSON(nil, err)
		}
		return err
	}

	if wantsJSON {
		return writeShowHowtoJSON(&shownHowto{
			Title:       howto.Name(),
			RawMarkdown: string(markdown),
		}, nil)
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

// Work around a limitation in how urfave/cli processes arguments and flags.
//
// If an argument starting with `-`, e.g., `-1` comes prior to the flags, the
// flags may not be processed correctly.  So the following both work:
//
//	show --format json -1
//	show 1 --format json
//
// but not this
//
//	show -1 --format json
//
// This function jumps through hoops to make the latter work.
func wantsJSONOutput(cmd *cli.Command) bool {
	if cmd.String("format") == "json" {
		return true
	}

	format := ""
	for index := 0; index < len(os.Args); index++ {
		arg := os.Args[index]
		if value, ok := strings.CutPrefix(arg, "--format="); ok {
			format = value
			continue
		}
		if arg == "--format" && index+1 < len(os.Args) {
			format = os.Args[index+1]
			index++
		}
	}
	return format == "json"
}

func loadHowtos(dirPath string, user *user.User) ([]*Howto, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	var howtos []*Howto
	for _, entry := range entries {
		filePath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			subFiles, err := loadHowtos(filePath, user)
			if err != nil {
				return nil, err
			}
			howtos = append(howtos, subFiles...)
			continue
		}

		if filepath.Ext(entry.Name()) == ".md" {
			relPath, err := filepath.Rel(howtoDir, filePath)
			if err != nil {
				return nil, err
			}
			howto := &Howto{Path: relPath}
			if (user == nil || user.Uid == "0") || !howto.IsAdminOnly() {
				howtos = append(howtos, howto)
			}
		}
	}
	sort.Sort(ByPath(howtos))
	return howtos, nil
}

func entriesTable(howtos []*Howto) error {
	namecolWidth := 7
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(cliui.AlcesBlue)).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			switch {
			case row == table.HeaderRow:
				return cliui.TableHeaderStyle
			case row%2 == 0:
				style = cliui.TableEvenRowStyle
			default:
				style = cliui.TableOddRowStyle
			}
			switch col {
			case 0:
				return style.Width(namecolWidth)
			}
			return style
		}).
		Width(termWidth)
	t.Headers("Index", "Title")
	for index, howto := range howtos {
		id := strconv.Itoa(index + 1)
		namecolWidth = max(namecolWidth, len(id)+2)
		titleColumn := lipgloss.JoinVertical(
			lipgloss.Left,
			howto.Name(),
		)
		t.Row(id, titleColumn)
	}
	_, err := lipgloss.Println(t)
	return err
}

type listedHowto struct {
	Index int    `json:"index"`
	Title string `json:"title"`
}

type listHowtoResponse struct {
	Success bool          `json:"success"`
	Guides  []listedHowto `json:"guides"`
	Error   string        `json:"error,omitempty"`
	Reason  string        `json:"reason,omitempty"`
}

type showHowtoResponse struct {
	Success bool       `json:"success"`
	Guide   shownHowto `json:"guide"`
	Error   string     `json:"error,omitempty"`
	Reason  string     `json:"reason,omitempty"`
}

type shownHowto struct {
	Title       string `json:"title"`
	RawMarkdown string `json:"raw_markdown"`
}

func writeListHowtosJSON(howtos []*Howto) error {
	guides := make([]listedHowto, 0, len(howtos))
	for index, howto := range howtos {
		guides = append(guides, listedHowto{
			Index: index + 1,
			Title: howto.Name(),
		})
	}
	response := listHowtoResponse{
		Success: true,
		Guides:  guides,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(response)
}

func writeListHowtosJSONError(listErr error) error {
	response := listHowtoResponse{
		Success: false,
		Guides:  []listedHowto{},
		Error:   listErr.Error(),
		Reason:  "unexpected",
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(response); err != nil {
		return err
	}
	return SilentExitError{
		ExitCode:  1,
		exitError: errors.New("listing howtos failed"),
	}
}

func writeShowHowtoJSON(guide *shownHowto, showErr error) error {
	if showErr != nil {
		response := showHowtoResponse{
			Success: false,
			Guide:   shownHowto{},
			Error:   showErr.Error(),
			Reason:  "unexpected",
		}
		if _, ok := errors.AsType[InvalidIndex](showErr); ok {
			response.Reason = "invalid_index"
		} else if _, ok := errors.AsType[UnknownHowto](showErr); ok {
			response.Reason = "not_found"
		}
		return writeHowtoShowResponse(response, 1)
	}

	response := showHowtoResponse{
		Success: true,
		Guide:   *guide,
	}
	return writeHowtoShowResponse(response, 0)
}

func writeHowtoShowResponse(response showHowtoResponse, exitCode int) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(response); err != nil {
		return err
	}
	if exitCode == 0 {
		return nil
	}
	return SilentExitError{
		ExitCode:  exitCode,
		exitError: errors.New("showing howto failed"),
	}
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
			if wantsJSONOutput(cmd) {
				return ctx, writeShowHowtoJSON(nil, MissingArguments{Args: missing})
			}
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

type InvalidIndex struct {
	Input string
}

func (ii InvalidIndex) Error() string {
	return fmt.Sprintf(
		"invalid input: '%s' is not a valid guide index. Use `flight howto list` to view the index for each user guide.",
		ii.Input,
	)
}

type SilentExitError struct {
	ExitCode  int
	exitError error
}

func (ee SilentExitError) Error() string {
	return ee.exitError.Error()
}
