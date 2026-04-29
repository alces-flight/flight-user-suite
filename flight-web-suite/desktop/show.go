package desktop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

func ShowCommand(ctx context.Context, env configenv.Env, username, sessionName string) (*Session, error) {
	cmd, err := buildDesktopCommand(ctx, env, username, "show", "--format", "json", sessionName)
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() != 0 {
			return nil, fmt.Errorf("showing desktop session: %s", stderr.String())
		}
		return nil, fmt.Errorf("showing desktop session: %w", err)
	}

	var session *Session
	if err := json.Unmarshal(stdout.Bytes(), &session); err != nil {
		return nil, fmt.Errorf("decoding desktop session: %w", err)
	}
	return session, nil
}
