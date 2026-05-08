package howto

import (
	"context"

	"github.com/concertim/flight-user-suite/flight-web-suite/desktop"
)

type GuideSummary struct {
	Index int    `json:"index"`
	Title string `json:"title"`
}

type ListResponse struct {
	Success bool           `json:"success"`
	Guides  []GuideSummary `json:"guides"`
	Error   string         `json:"error"`
	Reason  string         `json:"reason"`
}

func (cli *HowtoCli) ListCommand(ctx context.Context, username string) (ListResponse, error) {
	cli.Logger.Info("HOWTO", "action", "list", "username", username, "remote", false)
	cmd, err := cli.buildHowtoCommand(ctx, username, "list", "--format", "json")
	if err != nil {
		return ListResponse{}, err
	}
	response, err := desktop.RunLocal[ListResponse]("listing howtos", cmd)
	if err != nil {
		return ListResponse{}, err
	}
	if response.Guides == nil {
		response.Guides = []GuideSummary{}
	}
	return response, nil
}
