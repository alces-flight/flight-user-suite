package howto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

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

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	var response ListResponse
	decodeErr := json.Unmarshal(stdout.Bytes(), &response)
	if decodeErr == nil {
		if response.Guides == nil {
			response.Guides = []GuideSummary{}
		}
		return response, nil
	}

	if runErr != nil {
		if stderr.Len() != 0 {
			return ListResponse{}, fmt.Errorf("running howto list command: %s", stderr.String())
		}
		return ListResponse{}, fmt.Errorf("running howto list command: %w", runErr)
	}
	return ListResponse{}, fmt.Errorf("decoding howto list response: %w", decodeErr)
}
