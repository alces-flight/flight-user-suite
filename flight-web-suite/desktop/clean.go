package desktop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type cleanResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error"`
	Reason      string `json:"reason"`
}

type cleanCommandDocument struct {
	Success bool            `json:"success"`
	Results []cleanResponse `json:"results"`
}

func CleanCommand(ctx context.Context, env configenv.Env, username, sessionName string) (cleanResponse, error) {
	cmd, err := buildDesktopCommand(ctx, env, username, "clean", "--format", "json", "--", sessionName)
	if err != nil {
		return cleanResponse{}, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	var response cleanCommandDocument
	if decodeErr := json.Unmarshal(stdout.Bytes(), &response); decodeErr == nil {
		if len(response.Results) == 1 {
			return response.Results[0], nil
		}
		return cleanResponse{}, fmt.Errorf("decoding desktop clean response: expected 1 result, got %d", len(response.Results))
	}

	if runErr != nil {
		if stderr.Len() != 0 {
			return cleanResponse{}, fmt.Errorf("cleaning desktop session: %s", stderr.String())
		}
		return cleanResponse{}, fmt.Errorf("cleaning desktop session: %w", runErr)
	}
	return cleanResponse{}, fmt.Errorf("decoding desktop clean response: %s", stdout.String())
}
