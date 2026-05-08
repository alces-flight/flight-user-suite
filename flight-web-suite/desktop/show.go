package desktop

import (
	"context"
	"log/slog"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type showResponse struct {
	Success bool    `json:"success"`
	Session Session `json:"session"`
	Error   string  `json:"error"`
	Reason  string  `json:"reason"`
}

func ShowCommand(ctx context.Context, logger *slog.Logger, env configenv.Env, username, sessionName string) (showResponse, error) {
	logger.Info("DESKTOP SESSION", "action", "show", "name", sessionName, "username", username, "remote", false)
	cmd, err := buildLocalDesktopCommand(ctx, env, username, "show", "--format", "json", sessionName)
	if err != nil {
		return showResponse{}, err
	}
	response, err := RunLocal[showResponse]("showing desktop session", cmd)
	if err != nil {
		return showResponse{}, err
	}
	// TODO:
	// * If session has websockify pid of 0 and is either active or remote.
	// Webify it, locally or remotely as appropriate.
	return response, nil
}
