package howto

import (
	"context"

	"github.com/concertim/flight-user-suite/flight-web-suite/cli"
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

func (hcli *HowtoCli) ListCommand(ctx context.Context, username string) (ListResponse, error) {
	hcli.Logger().Info("HOWTO", "action", "list", "username", username, "remote", false)
	response, err := cli.RunLocal[ListResponse](ctx, "listing howtos", hcli, username, "list", "--format", "json")
	if err != nil {
		return ListResponse{}, err
	}
	if response.Guides == nil {
		response.Guides = []GuideSummary{}
	}
	return response, nil
}
