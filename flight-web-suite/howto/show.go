package howto

import (
	"context"
	"strconv"

	"github.com/concertim/flight-user-suite/flight-web-suite/cli"
)

type Guide struct {
	Title       string `json:"title"`
	RawMarkdown string `json:"raw_markdown"`
}

type ShowResponse struct {
	Success bool   `json:"success"`
	Guide   Guide  `json:"guide"`
	Error   string `json:"error"`
	Reason  string `json:"reason"`
}

func (hcli *HowtoCli) ShowCommand(ctx context.Context, username string, index int) (ShowResponse, error) {
	hcli.Logger().Info("HOWTO", "action", "show", "index", index, "username", username, "remote", false)
	return cli.RunLocal[ShowResponse](ctx, "showing howto", hcli, username, "show", "--format", "json", strconv.Itoa(index))
}
