package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"charm.land/log/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func webifySessionCommand() *cli.Command {
	return &cli.Command{
		Name:        "webify",
		Usage:       "Start web access support for an active desktop session",
		Description: wordwrap.String("Start required and optional web support programs for an active interactive desktop session. Can be used to add web support to sessions that were started before web dependencies were installed.", maxTextWidth),
		Category:    "Sessions",
		Flags:       []cli.Flag{formatFlag},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "name", UsageText: "<name>"},
		},
		Before:        assertArgPresent("name"),
		ShellComplete: completeActiveSessionNames,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.StringArg("name")
			session, err := loadSession(name)
			if cmd.String("format") == "json" {
				return webifySessionJSON(ctx, session, err)
			}
			if err != nil {
				if err2 := session.RemoveSessionDir(); err2 != nil {
					log.Debug("Removing session dir", "sessionDir", session.sessionDir(), "err", err2)
				}
				return err
			}
			switch session.ComputedState() {
			case "active":
				p := createPin("Starting web access support...")
				cancel := p.Start(ctx)
				defer cancel()
				timer := time.After(1 * time.Second)
				err = session.Webify(ctx)
				<-timer
				if err != nil {
					p.Fail("Starting web access support failed")
					return err
				}
				p.Stop("Web access support is ready!")
				return nil

			case "remote":
				return fmt.Errorf("session %s is not local", session.Name)
			default:
				return fmt.Errorf("session %s is not active", session.Name)
			}
		},
	}
}

type webifyResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

func webifySessionJSON(ctx context.Context, session *Session, loadErr error) error {
	if loadErr != nil {
		if _, ok := errors.AsType[UnknownSession](loadErr); ok {
			return writeWebifyFailure(session.Name, loadErr.Error(), "not_found")
		}
		return writeWebifyFailure(session.Name, loadErr.Error(), "webify_failed")
	}
	switch session.ComputedState() {
	case Remote:
		return writeWebifyFailure(session.Name, fmt.Sprintf("Desktop session '%s' is not local.", session.Name), "not_local")
	case Active:
		if err := session.Webify(ctx); err != nil {
			return writeWebifyFailure(session.Name, fmt.Sprintf("Starting web access support failed: %s", err.Error()), "webify_failed")
		}
		return writeWebifySuccess(session.Name)
	default:
		return writeWebifyFailure(session.Name, fmt.Sprintf("Desktop session '%s' is not active.", session.Name), "not_active")
	}
}

func writeWebifySuccess(sessionName string) error {
	return writeWebifyResponse(webifyResponse{
		Success:     true,
		SessionName: sessionName,
	}, 0)
}

func writeWebifyFailure(sessionName, message, reason string) error {
	return writeWebifyResponse(webifyResponse{
		Success:     false,
		SessionName: sessionName,
		Error:       message,
		Reason:      reason,
	}, 1)
}

func writeWebifyResponse(response webifyResponse, exitCode int) error {
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
		exitError: errors.New("session webification failed"),
	}
}
