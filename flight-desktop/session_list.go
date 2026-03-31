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
			return style
		}).
		Width(termWidth)
	t.Headers("Name", "Type", "Connection string", "Password", "State", "ID")
	for _, s := range sessions {
		t.Row(
			s.Name,
			s.SessionType,
			s.PrimaryConnectionString(),
			s.Password,
			string(s.SessionState),
			s.ID,
		)
	}
	_, err := lipgloss.Println(t)
	return err
}
