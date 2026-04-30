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

func showSessionCommand() *cli.Command {
	return &cli.Command{
		Name:        "show",
		Usage:       "Show information about a desktop session",
		Description: wordwrap.String("Display the connection information for a desktop session.", maxTextWidth),
		Category:    "Sessions",
		Flags:       []cli.Flag{formatFlag},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "name", UsageText: "<name>"},
		},
		// TODO: Move check inside action.
		Before:        assertArgPresent("name"),
		ShellComplete: completeSessionNames,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.StringArg("name")
			session, err := loadSession(name)
			if cmd.String("format") == "json" {
				return writeSessionJSON(session, err)
			}
			if err != nil {
				if err2 := session.RemoveSessionDir(); err2 != nil {
					log.Debug("Removing session dir", "sessionDir", session.sessionDir(), "err", err2)
				}
				return err
			}
			sessionInfo(session)
			if session.State != Exited && session.State != Broken {
				connectionInfo(session)
			}
			managementInfo(session)
			return nil
		},
	}
}

type showResponse struct {
	Success bool         `json:"success"`
	Session shownSession `json:"session,omitzero"`
	Error   string       `json:"error,omitempty"`
	Reason  string       `json:"reason,omitempty"`
}

type shownSession struct {
	Name          string       `json:"name"`
	DesktopType   string       `json:"desktop_type"`
	State         sessionState `json:"state"`
	Host          string       `json:"host"`
	Port          int          `json:"port"`
	WebsocketPort int          `json:"websocket_port"`
	Password      string       `json:"password"`
	CreatedAt     string       `json:"created_at"`
}

func writeSessionJSON(session *Session, loadErr error) error {
	if loadErr != nil {
		var response showResponse
		if _, ok := errors.AsType[UnknownSession](loadErr); ok {
			response = showResponse{
				Success: false,
				Session: shownSession{},
				Error:   loadErr.Error(),
				Reason:  "not_found",
			}
		} else {
			response = showResponse{
				Success: false,
				Session: shownSession{},
				Error:   loadErr.Error(),
				Reason:  "unexpected",
			}
		}
		return writeShowResponse(response, 1)
	}
	shownSession := shownSession{
		Name:          session.Name,
		DesktopType:   session.SessionType,
		State:         session.ComputedState(),
		Host:          session.Metadata.Host,
		Password:      session.Password,
		Port:          session.Port(),
		WebsocketPort: session.GetWebsocketPort(),
		CreatedAt:     session.CreatedAt.Format(time.RFC3339),
	}
	response := showResponse{
		Success: true,
		Session: shownSession,
		Error:   "",
		Reason:  "",
	}
	return writeShowResponse(response, 0)
}

func writeShowResponse(response showResponse, exitCode int) error {
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
		exitError: errors.New("showing session failed"),
	}
}

func completeSessionNames(ctx context.Context, cmd *cli.Command) {
	switch cmd.NArg() {
	case 0:
		sessions, err := loadAllSessions()
		if err != nil {
			return
		}
		for _, session := range sessions {
			fmt.Println(session.Name)
		}
	}
}
