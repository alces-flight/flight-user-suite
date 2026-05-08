package desktop

import (
	"context"

	"github.com/concertim/flight-user-suite/flight-web-suite/cli"
)

type showResponse struct {
	Success bool    `json:"success"`
	Session Session `json:"session"`
	Error   string  `json:"error"`
	Reason  string  `json:"reason"`
}

func (dcli *DesktopCli) ShowCommand(ctx context.Context, username, sessionName string) (showResponse, error) {
	dcli.logger.Info("DESKTOP SESSION", "action", "show", "name", sessionName, "username", username, "remote", false)
	response, err := cli.RunLocal[showResponse](ctx, "showing desktop session", dcli, username, "show", "--format", "json", sessionName)
	if err != nil {
		return showResponse{}, err
	}
	// TODO:
	// * If session has websockify pid of 0 and is either active or remote.
	// Webify it, locally or remotely as appropriate.
	return response, nil
}
