package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/concertim/flight-user-suite/flight/cliui"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func listSessionsCommand() *cli.Command {
	return &cli.Command{
		Name:        "list",
		Usage:       "List interactive desktop sessions",
		Description: wordwrap.String("Display all known desktop sessions and their states.", 80),
		Category:    "Sessions",
		Flags:       []cli.Flag{formatFlag},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			sessions, err := loadAllSessions()
			if err != nil {
				return err
			}
			if cmd.String("format") == "json" {
				return writeSessionsJSON(sessions)
			}
			if len(sessions) == 0 {
				fmt.Println("No desktop sessions found.")
				return nil
			}
			return sessionsTable(sessions)
		},
	}
}

type listedSession struct {
	Name        string       `json:"name"`
	DesktopType string       `json:"desktop_type"`
	State       sessionState `json:"state"`
	Host        string       `json:"host"`
	CreatedAt   string       `json:"created_at"`
}

func writeSessionsJSON(sessions []*Session) error {
	output := make([]listedSession, 0, len(sessions))
	for _, session := range sessions {
		output = append(output, listedSession{
			Name:        session.Name,
			DesktopType: session.SessionType,
			State:       session.ComputedState(),
			Host:        session.Metadata.Host,
			CreatedAt:   session.CreatedAt.Format(time.RFC3339),
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func sessionsTable(sessions []*Session) error {
	namecolWidth := 8
	maxNamecolWidth := 40
	typecolWidth := 8
	connectioncolWidth := 8
	passwordcolWidth := 10

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
			string(s.ComputedState()),
		)
	}
	_, err := lipgloss.Println(t)
	return err
}
