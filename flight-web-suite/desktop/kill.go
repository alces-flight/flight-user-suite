package desktop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type terminationResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error"`
	Reason      string `json:"reason"`
}

func KillCommand(ctx context.Context, env configenv.Env, username, sessionName string) (terminationResponse, error) {
	cmd, err := buildDesktopCommand(ctx, env, username, "kill", "--format", "json", "--", sessionName)
	if err != nil {
		return terminationResponse{}, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	var response terminationResponse
	if decodeErr := json.Unmarshal(stdout.Bytes(), &response); decodeErr == nil {
		return response, nil
	}

	if runErr != nil {
		if stderr.Len() != 0 {
			return terminationResponse{}, fmt.Errorf("terminating desktop session: %s", stderr.String())
		}
		return terminationResponse{}, fmt.Errorf("terminating desktop session: %w", runErr)
	}
	return terminationResponse{}, fmt.Errorf("decoding desktop termination response: %s", stdout.String())
}
