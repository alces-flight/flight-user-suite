package howto

import (
	"context"
	"strconv"

	"github.com/concertim/flight-user-suite/flight-web-suite/desktop"
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

func (cli *HowtoCli) ShowCommand(ctx context.Context, username string, index int) (ShowResponse, error) {
	cli.Logger.Info("HOWTO", "action", "show", "index", index, "username", username, "remote", false)
	cmd, err := cli.buildHowtoCommand(ctx, username, "show", "--format", "json", strconv.Itoa(index))
	if err != nil {
		return ShowResponse{}, err
	}
	return desktop.RunLocal[ShowResponse]("showing howto", cmd)
}
