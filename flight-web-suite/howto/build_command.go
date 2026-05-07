package howto

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/concertim/flight-user-suite/flight-web-suite/desktop"
	"github.com/concertim/flight-user-suite/flight/configenv"
)

func howtoToolPath(env configenv.Env) string {
	return filepath.Join(env.FlightRoot, "usr", "lib", "flight-core", "flight-howto")
}

func buildHowtoCommand(ctx context.Context, env configenv.Env, username string, args ...string) (*exec.Cmd, error) {
	cmd, err := desktop.BuildLocalCommand(ctx, howtoToolPath(env), username, args...)
	if err != nil {
		return nil, fmt.Errorf("building howto command: %w", err)
	}
	return cmd, nil
}
