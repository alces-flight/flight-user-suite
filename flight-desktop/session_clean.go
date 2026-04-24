package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/concertim/flight-user-suite/flight/cliui"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
	"github.com/yarlson/pin"
)

func cleanSessionCommand() *cli.Command {
	return &cli.Command{
		Name:  "clean",
		Usage: "Clean up one or more exited or broken desktop sessions",
		Description: wordwrap.String(`Remove one or more desktop session directories for exited or broken desktop sessions.

Depending on how cleanly desktop sessions exit, the session directory may be retained and require manual cleaning.  Desktops that are cleanly exited or manually terminated using the 'kill' command are automatically cleaned when the exit.

You can specify which sessions are cleaned by providing the optional <id> parameter one or more times. If no <id>s are specified, data for all sessions that are marked as 'exited' or 'broken' will be removed.`, maxTextWidth),

		Category: "Sessions",
		Arguments: []cli.Argument{
			&cli.StringArgs{Name: "id", UsageText: "[<id>...]", Min: 0, Max: -1},
		},
		Flags:         []cli.Flag{formatFlag},
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

			if cmd.String("format") == "json" {
				return cleanSessionJSON(sessions, len(ids) > 0)
			}
			return cleanSessionPretty(ctx, sessions)
		},
	}
}

func cleanSessionPretty(ctx context.Context, sessions []*Session) error {
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
	skippedStyle := lipgloss.NewStyle().Foreground(cliui.LightDark(cliui.Primary, cliui.Cream)).MarginLeft(1)
	cleanedStyle := lipgloss.NewStyle().Foreground(lipgloss.Green).MarginLeft(1)
	failedStyle := lipgloss.NewStyle().Foreground(lipgloss.Red).MarginLeft(1)
	cleaned := make([]*Session, 0)
	failed := make([]*Session, 0)
	skipped := make([]*Session, 0)

	for _, session := range sessions {
		if session.ComputedState() == Remote || session.ComputedState() == Active {
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
			append([]string{out, cliui.Header.Render("Cleaned sessions")}, coloured...)...,
		)
	}
	if len(failed) > 0 {
		coloured := make([]string, 0, len(failed))
		for _, session := range failed {
			coloured = append(coloured, failedStyle.Render(session.Name))
		}
		out = lipgloss.JoinVertical(
			lipgloss.Left,
			append([]string{out, cliui.Header.Render("Failed to clean")}, coloured...)...,
		)
	}
	if len(skipped) > 0 {
		coloured := make([]string, 0, len(skipped))
		for _, session := range skipped {
			coloured = append(coloured, skippedStyle.Render(session.Name))
		}
		out = lipgloss.JoinVertical(
			lipgloss.Left,
			append([]string{out, cliui.Header.Render("Skipped active/remote sessions")}, coloured...)...,
		)
	}

	if firstErr != nil {
		p.Fail("Cleaning failed")
	} else {
		p.Stop("Cleaning complete")
	}
	lipgloss.Println(out)
	return firstErr
}

type cleanResult struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type cleanCommandResponse struct {
	Success bool          `json:"success"`
	Results []cleanResult `json:"results"`
}

func cleanSessionJSON(sessions []*Session, namedSessionsOnly bool) error {
	results := make([]cleanResult, 0, len(sessions))
	allSucceeded := true

	if namedSessionsOnly {
		// We're cleaning an explicitly named set of sessions.
		for _, session := range sessions {
			result := cleanSingleSession(session)
			results = append(results, result)
			if !result.Success {
				allSucceeded = false
			}
		}
	} else {
		// We're cleaning all sessions that can be cleaned.
		for _, session := range sessions {
			if session.ComputedState() == Remote || session.ComputedState() == Active {
				// No sessions have been explicitly specified - we're cleaning
				// all possible sessions. Remote and active sessions should be
				// silently ignored.
				continue
			}
			result := cleanSingleSession(session)
			results = append(results, result)
			if !result.Success {
				allSucceeded = false
			}
		}
	}

	return writeCleanResponse(cleanCommandResponse{
		Success: allSucceeded,
		Results: results,
	})
}

func cleanSingleSession(session *Session) cleanResult {
	// Guard against sessions we're not going to clean.
	//
	// We do not clean active sessions, and we cannot determine if a remote
	// session is active so we don't clean them either.

	switch session.ComputedState() {
	case Remote:
		return cleanResult{
			Success:     false,
			SessionName: session.Name,
			Error:       fmt.Sprintf("Desktop session '%s' is not local.", session.Name),
			Reason:      "not_local",
		}
	case Active:
		return cleanResult{
			Success:     false,
			SessionName: session.Name,
			Error:       fmt.Sprintf("Desktop session '%s' is active.", session.Name),
			Reason:      "active",
		}
	}

	if err := session.RemoveSessionDir(); err != nil {
		return cleanResult{
			Success:     false,
			SessionName: session.Name,
			Error:       fmt.Sprintf("Cleaning session '%s' failed.", session.Name),
			Reason:      "clean_failed",
		}
	}
	return cleanResult{
		Success:     true,
		SessionName: session.Name,
	}
}

func writeCleanResponse(response cleanCommandResponse) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(response); err != nil {
		return err
	}
	if response.Success {
		return nil
	}
	return SilentExitError{
		ExitCode:  1,
		exitError: errors.New("session cleaning failed"),
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
			if session.ComputedState() == Exited || session.ComputedState() == Broken {
				fmt.Println(session.Name)
			}
		}
	}
}
