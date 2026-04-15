package main

import (
	"context"
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/concertim/flight-user-suite/flight/pkg"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
	"github.com/yarlson/pin"
)

func cleanSessionCommand() *cli.Command {
	return &cli.Command{
		Name:  "clean",
		Usage: "Clean up one or more exited desktop sessions",
		Description: wordwrap.String(`Remove one or more desktop session directories for exited or broken desktop sessions.

Depending on how cleanly desktop sessions exit, the session directory may be retained and require manual cleaning.  Desktops that are cleanly exited or manually terminated using the 'kill' command are automatically cleaned when the exit.

You can specify which sessions are cleaned by providing the optional <id> parameter one or more times. If no <id>s are specified, data for all sessions that are marked as 'exited' or 'broken' will be removed.`, maxTextWidth),

		Category: "Sessions",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "id", UsageText: "[<id>...]", Min: 0, Max: -1},
		},
		ShellComplete: completeExitedSessionNames,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			ids := cmd.StringArgs("id")
			var sessions []*Session
			if len(ids) == 0 {
				var err error
				sessions, err = loadAllSessions()
				if err != nil {
					return err
				}
			} else {
				sessions = make([]*Session, 0, len(ids))
				for _, id := range ids {
					session, _ := loadSession(id)
					sessions = append(sessions, session)
				}
			}
			p := pin.New("Cleaning desktop sessions",
				pin.WithSpinnerColor(pin.ColorCyan),
				pin.WithTextColor(pin.ColorGreen),
				pin.WithDoneSymbol('\u2705'),
				pin.WithFailSymbol('\u274c'),
				pin.WithFailColor(pin.ColorRed),
			)
			cancel := p.Start(ctx)
			defer cancel()
			timer := time.After(1 * time.Second)

			var firstErr error
			skippedStyle := lipgloss.NewStyle().Foreground(pkg.LightDark(pkg.Primary, pkg.Cream)).MarginLeft(1)
			cleanedStyle := lipgloss.NewStyle().Foreground(lipgloss.Green).MarginLeft(1)
			failedStyle := lipgloss.NewStyle().Foreground(lipgloss.Red).MarginLeft(1)
			cleaned := make([]*Session, 0)
			failed := make([]*Session, 0)
			skipped := make([]*Session, 0)

			for _, session := range sessions {
				if !session.IsLocal() || session.State == Active {
					skipped = append(skipped, session)
				} else {
					err := session.RemoveSessionDir()
					if err != nil {
						failed = append(failed, session)
						if firstErr == nil {
							firstErr = err
						}
					} else {
						cleaned = append(cleaned, session)
					}
				}
			}

			<-timer

			var out string
			if len(cleaned) > 0 {
				coloured := make([]string, 0, len(cleaned))
				for _, session := range cleaned {
					coloured = append(coloured, cleanedStyle.Render(session.Name))
				}
				out = lipgloss.JoinVertical(
					lipgloss.Left,
					append([]string{out, pkg.Header.Render("Cleaned exited sessions")}, coloured...)...,
				)
			}
			if len(failed) > 0 {
				coloured := make([]string, 0, len(failed))
				for _, session := range failed {
					coloured = append(coloured, failedStyle.Render(session.Name))
				}
				out = lipgloss.JoinVertical(
					lipgloss.Left,
					append([]string{out, pkg.Header.Render("Failed to clean")}, coloured...)...,
				)
			}
			if len(skipped) > 0 {
				coloured := make([]string, 0, len(skipped))
				for _, session := range skipped {
					coloured = append(coloured, skippedStyle.Render(session.Name))
				}
				out = lipgloss.JoinVertical(
					lipgloss.Left,
					append([]string{out, pkg.Header.Render("Skipped active/remote sessions")}, coloured...)...,
				)
			}

			if firstErr != nil {
				p.Fail("Cleaning failed")
			} else {
				p.Stop("Cleaning complete")
			}
			lipgloss.Println(out)
			return firstErr
		},
	}
}

func completeExitedSessionNames(ctx context.Context, cmd *cli.Command) {
	switch cmd.NArg() {
	case 0:
		sessions, err := loadAllSessions()
		if err != nil {
			return
		}
		for _, session := range sessions {
			if session.SessionState() == Exited {
				fmt.Println(session.Name)
			}
		}
	}
}
