package desktop

import (
	"context"
	"fmt"
	"strings"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type cleanResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error"`
	Reason      string `json:"reason"`
}

type cleanCommandDocument struct {
	Success bool            `json:"success"`
	Results []cleanResponse `json:"results"`
}

func CleanCommand(ctx context.Context, env configenv.Env, username, sessionName string) (cleanResponse, error) {
	showResponse, err := ShowCommand(ctx, env, username, sessionName)
	if err != nil {
		return cleanResponse{}, err
	}
	if showResponse.Reason == "not_found" {
		// It's already gone.
		return cleanResponse{Success: true, SessionName: sessionName}, nil
	}
	args := []string{"clean", "--format", "json", "--", sessionName}
	if showResponse.Session.State == "remote" {
		return remoteClean(env, username, showResponse.Session, args)
	}
	return localClean(ctx, env, username, args)
}

func localClean(ctx context.Context, env configenv.Env, username string, args []string) (cleanResponse, error) {
	cmd, err := buildLocalDesktopCommand(ctx, env, username, args...)
	if err != nil {
		return cleanResponse{}, err
	}
	response, err := RunLocal[cleanCommandDocument]("cleaning desktop session", cmd)
	if err != nil {
		return cleanResponse{}, err
	}
	if len(response.Results) == 1 {
		return response.Results[0], nil
	}
	return cleanResponse{}, fmt.Errorf("decoding desktop clean response: expected 1 result, got %d", len(response.Results))
}

func remoteClean(env configenv.Env, username string, session Session, args []string) (cleanResponse, error) {
	cmd := append([]string{desktopToolPath(env)}, args...)
	cmdString := strings.Join(cmd, " ")
	return runRemoteCommand[cleanResponse]("cleaning desktop session", username, cmdString, session.Host)
}
