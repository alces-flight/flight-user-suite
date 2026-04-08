package main

import (
	"context"
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func listSessionsCommand() *cli.Command {
	return &cli.Command{
		Name:        "list",
		Usage:       "List interactive desktop sessions",
		Description: wordwrap.String("Display all known desktop sessions and their states.", 80),
		Category:    "Sessions",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			sessions, err := loadAllSessions()
			if err != nil {
				return err
			}
			if len(sessions) == 0 {
				fmt.Println("No desktop sessions found.")
				return nil
			}
			err = sessionsTable(sessions)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func sessionsTable(sessions []*Session) error {
	namecolWidth := 8
	maxNamecolWidth := 40
	typecolWidth := 8
	connectioncolWidth := 8
	passwordcolWidth := 10

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
			case 1:
				return style.Width(typecolWidth)
			case 2:
				return style.MaxWidth(connectioncolWidth)
			case 3:
				return style.Width(passwordcolWidth)
			case 4:
				return style.Width(8)
			}
			return style
		}).
		Width(termWidth)
	t.Headers("Name", "Type", "Connection string", "Password", "State")
	for _, s := range sessions {
		connectionString := s.PrimaryConnectionString()
		namecolWidth = min(max(namecolWidth, len(s.Name)+2), maxNamecolWidth)
		namecolWidth = max(namecolWidth, len(s.Name)+2)
		typecolWidth = max(typecolWidth, len(s.SessionType)+2)
		connectioncolWidth = max(connectioncolWidth, len(connectionString)+2)
		passwordcolWidth = max(passwordcolWidth, len(s.Password)+2)

		t.Row(
			s.Name,
			s.SessionType,
			connectionString,
			s.Password,
			string(s.SessionState()),
		)
	}
	_, err := lipgloss.Println(t)
	return err
}
