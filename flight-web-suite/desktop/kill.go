package desktop

import (
	"context"
	"strings"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type terminationResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error"`
	Reason      string `json:"reason"`
}

func KillCommand(ctx context.Context, env configenv.Env, username, sessionName string) (terminationResponse, error) {
	showResponse, err := ShowCommand(ctx, env, username, sessionName)
	if err != nil {
		return terminationResponse{}, err
	}
	if showResponse.Reason == "not_found" {
		// It's already gone.
		return terminationResponse{Success: true, SessionName: sessionName}, nil
	}
	args := []string{"kill", "--format", "json", "--", sessionName}
	if showResponse.Session.State == "remote" {
		return remoteKill(env, username, showResponse.Session, args)
	}
	return localKill(ctx, env, username, args)

}

func localKill(ctx context.Context, env configenv.Env, username string, args []string) (terminationResponse, error) {
	cmd, err := buildLocalDesktopCommand(ctx, env, username, args...)
	if err != nil {
		return terminationResponse{}, err
	}
	return RunLocal[terminationResponse]("terminating desktop session", cmd)
}

func remoteKill(env configenv.Env, username string, session Session, args []string) (terminationResponse, error) {
	cmd := append([]string{desktopToolPath(env)}, args...)
	cmdString := strings.Join(cmd, " ")
	return runRemoteCommand[terminationResponse]("terminating desktop session", username, cmdString, session.Host)
}
