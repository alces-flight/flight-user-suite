package main

import (
	"context"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/concertim/flight-user-suite/flight/pkg"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func statusCommand() *cli.Command {
	return &cli.Command{
		Name:        "status",
		Usage:       "Print the status of the Flight Web Suite service",
		Description: wordwrap.String("Prints the status of the Flight Web Suite service, including whether it is running and if so the address it is listening on.", maxTextWidth),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return statusTable()
		},
	}
}

func statusTable() error {
	namecolWidth := 8
	maxNamecolWidth := 40

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
			case 1:
				return style.Width(9)
			}
			return style
		}).
		Width(termWidth)
	t.Headers("Name", "Status", "URL")

	service := Service{ID: "web-suite", Name: "Web Suite"}
	namecolWidth = min(max(namecolWidth, len(service.Name)+2), maxNamecolWidth)
	namecolWidth = max(namecolWidth, len(service.Name)+2)
	t.Row(service.Name, service.State(), "-")

	_, err := lipgloss.Println(t)
	return err
}
