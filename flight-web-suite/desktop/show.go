package desktop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type showResponse struct {
	Success bool    `json:"success"`
	Session Session `json:"session"`
	Error   string  `json:"error"`
	Reason  string  `json:"reason"`
}

func ShowCommand(ctx context.Context, env configenv.Env, username, sessionName string) (showResponse, error) {
	cmd, err := buildDesktopCommand(ctx, env, username, "show", "--format", "json", sessionName)
	if err != nil {
		return showResponse{}, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	var response showResponse
	if decodeErr := json.Unmarshal(stdout.Bytes(), &response); decodeErr == nil {
		return response, nil
	}

	if runErr != nil {
		if stderr.Len() != 0 {
			return showResponse{}, fmt.Errorf("showing desktop session: %s", stderr.String())
		}
		return showResponse{}, fmt.Errorf("showing desktop session: %w", runErr)
	}
	return showResponse{}, fmt.Errorf("decoding desktop show response: %s", stdout.String())
}
