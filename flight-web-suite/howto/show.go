package howto

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/concertim/flight-user-suite/flight-web-suite/desktop"
	"github.com/concertim/flight-user-suite/flight/configenv"
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

func ShowCommand(ctx context.Context, logger *slog.Logger, env configenv.Env, username string, index int) (ShowResponse, error) {
	logger.Info("HOWTO", "action", "show", "index", index, "username", username, "remote", false)
	cmd, err := buildHowtoCommand(ctx, env, username, "show", "--format", "json", strconv.Itoa(index))
	if err != nil {
		return ShowResponse{}, err
	}
	return desktop.RunLocal[ShowResponse]("showing howto", cmd)
}
