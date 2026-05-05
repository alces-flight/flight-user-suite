package howto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

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

func ShowCommand(ctx context.Context, env configenv.Env, username string, index int) (ShowResponse, error) {
	cmd, err := buildHowtoCommand(ctx, env, username, "show", "--format", "json", strconv.Itoa(index))
	if err != nil {
		return ShowResponse{}, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	var response ShowResponse
	decodeErr := json.Unmarshal(stdout.Bytes(), &response)
	if decodeErr == nil {
		return response, nil
	}

	if runErr != nil {
		if stderr.Len() != 0 {
			return ShowResponse{}, fmt.Errorf("running howto show command: %s", stderr.String())
		}
		return ShowResponse{}, fmt.Errorf("running howto show command: %w", runErr)
	}
	return ShowResponse{}, fmt.Errorf("decoding howto show response: %w", decodeErr)
}
