package desktop

import (
	"log/slog"
	"path/filepath"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type DesktopCli struct {
	Env    configenv.Env
	Logger *slog.Logger
}

func (cli *DesktopCli) ToolPath() string {
	return filepath.Join(cli.Env.FlightRoot, "usr", "lib", "flight-core", "flight-desktop")
}
