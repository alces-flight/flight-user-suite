package desktop

import (
	"context"
)

type showResponse struct {
	Success bool    `json:"success"`
	Session Session `json:"session"`
	Error   string  `json:"error"`
	Reason  string  `json:"reason"`
}

func (cli *DesktopCli) ShowCommand(ctx context.Context, username, sessionName string) (showResponse, error) {
	cli.Logger.Info("DESKTOP SESSION", "action", "show", "name", sessionName, "username", username, "remote", false)
	cmd, err := cli.buildLocalDesktopCommand(ctx, username, "show", "--format", "json", sessionName)
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
