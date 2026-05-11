package desktop

import (
	"context"
	"fmt"

	"github.com/concertim/flight-user-suite/flight-web-suite/cli"
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

func (dcli *DesktopCli) CleanCommand(ctx context.Context, username, sessionName string) (cleanResponse, error) {
	showResponse, err := dcli.ShowCommand(ctx, username, sessionName, false)
	if err != nil {
		return cleanResponse{}, err
	}
	if showResponse.Reason == "not_found" {
		// It's already gone.
		dcli.logger.Info("DESKTOP SESSION", "action", "clean", "name", sessionName, "username", username, "err", "not found")
		return cleanResponse{Success: true, SessionName: sessionName}, nil
	}
	args := []string{"clean", "--format", "json", "--", sessionName}
	if showResponse.Session.State == "remote" {
		dcli.logger.Info("DESKTOP SESSION", "action", "clean", "name", sessionName, "username", username, "remote", true, "host", showResponse.Session.Host)
		return dcli.remoteClean(username, showResponse.Session.Host, args)
	}
	dcli.logger.Info("DESKTOP SESSION", "action", "clean", "name", sessionName, "username", username, "remote", false)
	return dcli.localClean(ctx, username, args)
}

func (dcli *DesktopCli) localClean(ctx context.Context, username string, args []string) (cleanResponse, error) {
	response, err := cli.RunLocal[cleanCommandDocument](ctx, "cleaning desktop session", dcli, username, args...)
	if err != nil {
		return cleanResponse{}, err
	}
	if len(response.Results) == 1 {
		return response.Results[0], nil
	}
	return cleanResponse{}, fmt.Errorf("decoding desktop clean response: expected 1 result, got %d", len(response.Results))
}

func (dcli *DesktopCli) remoteClean(username, host string, args []string) (cleanResponse, error) {
	response, err := cli.RunRemote[cleanCommandDocument]("cleaning desktop session", dcli, username, host, args)
	if err != nil {
		return cleanResponse{}, err
	}
	if len(response.Results) == 1 {
		return response.Results[0], nil
	}
	return cleanResponse{}, fmt.Errorf("decoding desktop clean response: expected 1 result, got %d", len(response.Results))
}
