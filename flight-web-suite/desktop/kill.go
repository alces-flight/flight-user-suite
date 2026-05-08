package desktop

import (
	"context"
	"log/slog"
	"strings"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type terminationResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error"`
	Reason      string `json:"reason"`
}

func KillCommand(ctx context.Context, logger *slog.Logger, env configenv.Env, username, sessionName string) (terminationResponse, error) {
	showResponse, err := ShowCommand(ctx, logger, env, username, sessionName)
	if err != nil {
		return terminationResponse{}, err
	}
	if showResponse.Reason == "not_found" {
		// It's already gone.
		logger.Info("DESKTOP SESSION", "action", "kill", "name", sessionName, "username", username, "err", "not found")
		return terminationResponse{Success: true, SessionName: sessionName}, nil
	}
	args := []string{"kill", "--format", "json", "--", sessionName}
	if showResponse.Session.State == "remote" {
		logger.Info("DESKTOP SESSION", "action", "kill", "name", sessionName, "username", username, "remote", true, "host", showResponse.Session.Host)
		return remoteKill(env, username, showResponse.Session.Host, args)
	}
	logger.Info("DESKTOP SESSION", "action", "kill", "name", sessionName, "username", username, "remote", false)
	return localKill(ctx, env, username, args)

}

func localKill(ctx context.Context, env configenv.Env, username string, args []string) (terminationResponse, error) {
	cmd, err := buildLocalDesktopCommand(ctx, env, username, args...)
	if err != nil {
		return terminationResponse{}, err
	}
	return RunLocal[terminationResponse]("terminating desktop session", cmd)
}

func remoteKill(env configenv.Env, username, host string, args []string) (terminationResponse, error) {
	cmd := append([]string{desktopToolPath(env)}, args...)
	cmdString := strings.Join(cmd, " ")
	return runRemoteCommand[terminationResponse]("terminating desktop session", username, cmdString, host)
}
