package desktop

import (
	"context"
	"fmt"
	"strings"
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

func (cli *DesktopCli) CleanCommand(ctx context.Context, username, sessionName string) (cleanResponse, error) {
	showResponse, err := cli.ShowCommand(ctx, username, sessionName)
	if err != nil {
		return cleanResponse{}, err
	}
	if showResponse.Reason == "not_found" {
		// It's already gone.
		cli.Logger.Info("DESKTOP SESSION", "action", "clean", "name", sessionName, "username", username, "err", "not found")
		return cleanResponse{Success: true, SessionName: sessionName}, nil
	}
	args := []string{"clean", "--format", "json", "--", sessionName}
	if showResponse.Session.State == "remote" {
		cli.Logger.Info("DESKTOP SESSION", "action", "clean", "name", sessionName, "username", username, "remote", true, "host", showResponse.Session.Host)
		return cli.remoteClean(username, showResponse.Session.Host, args)
	}
	cli.Logger.Info("DESKTOP SESSION", "action", "clean", "name", sessionName, "username", username, "remote", false)
	return cli.localClean(ctx, username, args)
}

func (cli *DesktopCli) localClean(ctx context.Context, username string, args []string) (cleanResponse, error) {
	cmd, err := cli.buildLocalDesktopCommand(ctx, username, args...)
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

func (cli *DesktopCli) remoteClean(username, host string, args []string) (cleanResponse, error) {
	cmd := append([]string{cli.ToolPath()}, args...)
	cmdString := strings.Join(cmd, " ")
	return runRemoteCommand[cleanResponse]("cleaning desktop session", username, cmdString, host, cli.Config)
}
