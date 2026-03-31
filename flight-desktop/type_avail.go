package main

import (
	"context"
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
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
		BorderStyle(lipgloss.NewStyle().Foreground(ctmOrange)).
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
			}
			return style
		}).
		Width(termWidth)
	t.Headers("Name", "Summary")
	for _, typ := range types {
		namecolWidth = max(namecolWidth, len(typ.Name)+2)
		summary := lipgloss.JoinVertical(
			lipgloss.Left,
			typ.Summary,
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				lipgloss.NewStyle().MarginRight(1).Render(">"),
				hyperlink.MarginBottom(1).Hyperlink(typ.URL).Render(typ.URL),
			),
		)
		t.Row(typ.Name, summary)
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
