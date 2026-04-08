package main

import (
	"context"
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"charm.land/log/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func typeAvailCommand() *cli.Command {
	return &cli.Command{
		Name:        "avail",
		Usage:       "Show available desktop types",
		Description: wordwrap.String("Display a list of available desktop types.", maxTextWidth),
		Category:    "Desktop types",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			types, err := loadAllTypes()
			if err != nil {
				return err
			}
			if len(types) == 0 {
				fmt.Println("No desktop types found.")
				return nil
			}
			err = typesTable(types)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func typesTable(types []*Type) error {
	namecolWidth := 8

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(alcesBlue)).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			switch {
			case row == table.HeaderRow:
				return tableHeaderStyle
			case row%2 == 0:
				style = tableEvenRowStyle
			default:
				style = tableOddRowStyle
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
				hyperlink.MarginBottom(1).Hyperlink(typ.URL).Render(typ.URL),
			),
		)

		depsText := "\u2754 Unknown"
		depsLoadError := typ.loadDependencies()

		if depsLoadError == nil {
			_, depsOK := runDoctor(requiredDependencies(typ.dependencies))

			depsText = lipgloss.NewStyle().Foreground(lipgloss.Red).Render("\u274c Missing")
			if depsOK {
				depsText = lipgloss.NewStyle().Foreground(lipgloss.Green).Render("\u2705 OK")
			}
		} else {
			log.Debug("While loading dependencies for type", "type", typ.ID, "err", depsLoadError)
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
