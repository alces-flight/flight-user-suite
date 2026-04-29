package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"charm.land/log/v2"
	"github.com/concertim/flight-user-suite/flight/cliui"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func typeAvailCommand() *cli.Command {
	return &cli.Command{
		Name:        "avail",
		Usage:       "Show available desktop types",
		Description: wordwrap.String("Display a list of available desktop types.", maxTextWidth),
		Category:    "Desktop types",
		Flags:       []cli.Flag{formatFlag},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			types, err := loadAllTypes(true)
			if err != nil {
				return err
			}
			slices.SortFunc(types, func(a, b *Type) int {
				return compareStrings(a.ID, b.ID)
			})
			if cmd.String("format") == "json" {
				return writeTypesJSON(types)
			}
			if len(types) == 0 {
				fmt.Println("No desktop types found.")
				return nil
			}
			return typesTable(types)
		},
	}
}

func writeTypesJSON(types []*Type) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(types)
}

func typesTable(types []*Type) error {
	namecolWidth := 8

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
			case 2:
				return style.Width(15)
			}
			return style
		}).
		Width(termWidth)
	t.Headers("Name", "Summary", "Dependencies")
	for _, typ := range types {
		namecolWidth = max(namecolWidth, len(typ.ID)+2)
		summary := lipgloss.JoinVertical(
			lipgloss.Left,
			typ.Summary,
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				lipgloss.NewStyle().MarginRight(1).Render(">"),
				cliui.Hyperlink.MarginBottom(1).Hyperlink(typ.URL).Render(typ.URL),
			),
		)

		depsText := "\u2754 Unknown"

		if typ.dependenciesLoadError == nil {
			depsText = lipgloss.NewStyle().Foreground(lipgloss.Red).Render("\u274c Missing")
			if typ.IsAvailable {
				depsText = lipgloss.NewStyle().Foreground(lipgloss.Green).Render("\u2705 OK")
			}
		} else {
			log.Debug("While loading dependencies for type", "type", typ.ID, "err", typ.dependenciesLoadError)
		}

		t.Row(typ.ID, summary, depsText)
	}
	_, err := lipgloss.Println(t)
	return err
}

type UnknownType struct {
	Type string
}

func (ut UnknownType) Error() string {
	return fmt.Sprintf("Unknown desktop type: %s", ut.Type)
}

func compareStrings(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
