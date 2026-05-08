package desktop

import (
	"context"
	"strings"
)

type terminationResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error"`
	Reason      string `json:"reason"`
}

func (cli *DesktopCli) KillCommand(ctx context.Context, username, sessionName string) (terminationResponse, error) {
	showResponse, err := cli.ShowCommand(ctx, username, sessionName)
	if err != nil {
		return terminationResponse{}, err
	}
	if showResponse.Reason == "not_found" {
		// It's already gone.
		cli.Logger.Info("DESKTOP SESSION", "action", "kill", "name", sessionName, "username", username, "err", "not found")
		return terminationResponse{Success: true, SessionName: sessionName}, nil
	}
	args := []string{"kill", "--format", "json", "--", sessionName}
	if showResponse.Session.State == "remote" {
		cli.Logger.Info("DESKTOP SESSION", "action", "kill", "name", sessionName, "username", username, "remote", true, "host", showResponse.Session.Host)
		return cli.remoteKill(username, showResponse.Session.Host, args)
	}
	cli.Logger.Info("DESKTOP SESSION", "action", "kill", "name", sessionName, "username", username, "remote", false)
	return cli.localKill(ctx, username, args)

}

func (cli *DesktopCli) localKill(ctx context.Context, username string, args []string) (terminationResponse, error) {
	cmd, err := cli.buildLocalDesktopCommand(ctx, username, args...)
	if err != nil {
		return terminationResponse{}, err
	}
	return RunLocal[terminationResponse]("terminating desktop session", cmd)
}

func (cli *DesktopCli) remoteKill(username, host string, args []string) (terminationResponse, error) {
	cmd := append([]string{cli.ToolPath()}, args...)
	cmdString := strings.Join(cmd, " ")
	return runRemoteCommand[terminationResponse]("terminating desktop session", username, cmdString, host)
}
