package howto

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"

	"github.com/concertim/flight-user-suite/flight-web-suite/desktop"
	"github.com/concertim/flight-user-suite/flight/configenv"
)

type HowtoCli struct {
	Env    configenv.Env
	Logger *slog.Logger
}

func (cli *HowtoCli) ToolPath() string {
	return filepath.Join(cli.Env.FlightRoot, "usr", "lib", "flight-core", "flight-howto")
}

func (cli *HowtoCli) buildHowtoCommand(ctx context.Context, username string, args ...string) (*exec.Cmd, error) {
	cmd, err := desktop.BuildLocalCommand(ctx, cli.ToolPath(), username, args...)
	if err != nil {
		return nil, fmt.Errorf("building howto command: %w", err)
	}
	return cmd, nil
}
