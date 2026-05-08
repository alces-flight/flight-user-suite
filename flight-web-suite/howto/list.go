package howto

import (
	"context"

	"github.com/concertim/flight-user-suite/flight-web-suite/desktop"
	"github.com/concertim/flight-user-suite/flight/configenv"
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

func ListCommand(ctx context.Context, env configenv.Env, username string) (ListResponse, error) {
	cmd, err := buildHowtoCommand(ctx, env, username, "list", "--format", "json")
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
