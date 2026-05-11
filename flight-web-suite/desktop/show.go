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

	if !webify || !response.Success || (response.Session.State != "active" && response.Session.State != "remote") {
		// We don't care if the session is webified.
		return response, nil

	} else if !response.Session.IsWebified {
		// We want the session to be webified and we know that it isn't.
		dcli.webify(ctx, username, sessionName, response.Session)
		return cli.RunLocal[showResponse](ctx, "showing desktop session", dcli, username, "show", "--format", "json", sessionName)

	} else if response.Session.State == "active" {
		// We want the session to be webified, it claims to be and we trust
		// that claim: its a local session.
		return response, nil
	}

	// We want the session to be webified, it claims to be but we don't trust
	// that claim: its a remote (aka non-local) session.
	//
	// Remote sessions claim to be webified, if they have *ever* been
	// webified.  Whereas, local sessions claim to be webified if they are
	// *currently* webified.
	//
	// Re-run the show command on the session's host, so that we can determine
	// if it is *currently* webified.
	origResponse := response
	response, err = cli.RunRemote[showResponse]("showing desktop session", dcli, username, response.Session.Host, "show", "--format", "json", sessionName)
	if err != nil || !response.Success {
		// Depending on why this failed, the correct response might be to take
		// the latest response as authoritative, or it might be to use the
		// first response as authoritative.  For instance, if `response.Reason`
		// is `not_found`, the session has likely been removed and this latest
		// response is authoritative.  However, if there is a transient
		// SSH/connection error, it might be better to take the first response
		// as authoritative (or not).
		return response, err
	}
	if response.Session.IsWebified {
		return response, nil
	}
	dcli.webify(ctx, username, sessionName, origResponse.Session)
	return cli.RunLocal[showResponse](ctx, "showing desktop session", dcli, username, "show", "--format", "json", sessionName)
}

func (dcli *DesktopCli) webify(ctx context.Context, username, sessionName string, session Session) {
	args := []string{"webify", "--format", "json", "--", sessionName}
	if session.State == "remote" {
		dcli.Logger().Info("DESKTOP SESSION", "action", "webify", "name", sessionName, "username", username, "remote", true, "host", session.Host)
		_, err := cli.RunRemote[webifyResponse]("webifying desktop session", dcli, username, session.Host, args...)
		if err != nil {
			dcli.Logger().Debug("DESKTOP SESSION", "action", "webify", "name", sessionName, "username", username, "remote", true, "host", session.Host, "error", err)
		}
	} else {
		dcli.Logger().Info("DESKTOP SESSION", "action", "webify", "name", sessionName, "username", username, "remote", false)
		_, err := cli.RunLocal[webifyResponse](ctx, "webifying desktop session", dcli, username, args...)
		if err != nil {
			dcli.Logger().Debug("DESKTOP SESSION", "action", "webify", "name", sessionName, "username", username, "remote", false, "error", err)
		}
	}
}
