package desktop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type Session struct {
	Name           string    `json:"name"`
	DesktopType    string    `json:"desktop_type"`
	State          string    `json:"state"`
	Host           string    `json:"host"`
	Port           int       `json:"port,omitempty"`
	WebsocketPort  int       `json:"websocket_port,omitempty"`
	Password       string    `json:"password,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	ScreenshotPath string    `json:"screenshot_path"`
	IsWebified     bool      `json:"is_webified"`
}

func ListCommand(ctx context.Context, logger *slog.Logger, env configenv.Env, username string) ([]*Session, error) {
	logger.Info("DESKTOP SESSION", "action", "list", "username", username, "remote", false)
	cmd, err := buildLocalDesktopCommand(ctx, env, username, "list", "--format", "json")
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() != 0 {
			return nil, fmt.Errorf("listing desktop sessions: %s", stderr.String())
		}
		return nil, fmt.Errorf("listing desktop sessions: %w", err)
	}

	var sessions []*Session
	if err := json.Unmarshal(stdout.Bytes(), &sessions); err != nil {
		return nil, fmt.Errorf("decoding desktop sessions: %w", err)
	}
	return sessions, nil
}
