package desktop

import (
	"context"

	"github.com/concertim/flight-user-suite/flight-web-suite/cli"
)

type terminationResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error"`
	Reason      string `json:"reason"`
}

func (dcli *DesktopCli) KillCommand(ctx context.Context, username, sessionName string) (terminationResponse, error) {
	showResponse, err := dcli.ShowCommand(ctx, username, sessionName)
	if err != nil {
		return terminationResponse{}, err
	}
	if showResponse.Reason == "not_found" {
		// It's already gone.
		dcli.logger.Info("DESKTOP SESSION", "action", "kill", "name", sessionName, "username", username, "err", "not found")
		return terminationResponse{Success: true, SessionName: sessionName}, nil
	}
	args := []string{"kill", "--format", "json", "--", sessionName}
	if showResponse.Session.State == "remote" {
		dcli.logger.Info("DESKTOP SESSION", "action", "kill", "name", sessionName, "username", username, "remote", true, "host", showResponse.Session.Host)
		return cli.RunRemote[terminationResponse]("terminating desktop session", dcli, username, showResponse.Session.Host, args)
	}
	dcli.logger.Info("DESKTOP SESSION", "action", "kill", "name", sessionName, "username", username, "remote", false)
	return cli.RunLocal[terminationResponse](ctx, "terminating desktop session", dcli, username, args...)

}
