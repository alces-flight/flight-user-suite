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

type webifyResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error"`
	Reason      string `json:"reason"`
}

func (dcli *DesktopCli) ShowCommand(ctx context.Context, username, sessionName string, webify bool) (showResponse, error) {
	dcli.logger.Info("DESKTOP SESSION", "action", "show", "name", sessionName, "username", username, "remote", false)
	response, err := cli.RunLocal[showResponse](ctx, "showing desktop session", dcli, username, "show", "--format", "json", sessionName)
	if err != nil {
		return showResponse{}, err
	}
	switch {
	case !webify,
		!response.Success,
		response.Session.State != "active" && response.Session.State != "remote",
		response.Session.IsWebified:
		return response, nil
	default:
		dcli.webify(ctx, username, sessionName, response)
		return cli.RunLocal[showResponse](ctx, "showing desktop session", dcli, username, "show", "--format", "json", sessionName)
	}
}

func (dcli *DesktopCli) webify(ctx context.Context, username, sessionName string, response showResponse) {
	args := []string{"webify", "--format", "json", "--", sessionName}
	if response.Session.State == "remote" {
		dcli.Logger().Info("DESKTOP SESSION", "action", "webify", "name", sessionName, "username", username, "remote", true, "host", response.Session.Host)
		_, err := cli.RunRemote[webifyResponse]("webifying desktop session", dcli, username, response.Session.Host, args)
		if err != nil {
			dcli.Logger().Debug("DESKTOP SESSION", "action", "webify", "name", sessionName, "username", username, "remote", true, "host", response.Session.Host, "error", err)
		}
	} else {
		dcli.Logger().Info("DESKTOP SESSION", "action", "webify", "name", sessionName, "username", username, "remote", false)
		_, err := cli.RunLocal[webifyResponse](ctx, "webifying desktop session", dcli, username, args...)
		if err != nil {
			dcli.Logger().Debug("DESKTOP SESSION", "action", "webify", "name", sessionName, "username", username, "remote", false, "error", err)
		}
	}
}
