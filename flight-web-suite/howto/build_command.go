package howto

import (
	"log/slog"
	"path/filepath"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type HowtoCli struct {
	env    configenv.Env
	logger *slog.Logger
}

func NewCliTool(logger *slog.Logger, env configenv.Env) *HowtoCli {
	return &HowtoCli{
		env:    env,
		logger: logger,
	}
}

func (hcli *HowtoCli) ToolPath() string {
	return filepath.Join(hcli.env.FlightRoot, "usr", "lib", "flight-core", "flight-howto")
}

func (hcli *HowtoCli) GetEnv() configenv.Env {
	return hcli.env
}

func (hcli *HowtoCli) Logger() *slog.Logger {
	return hcli.logger
}
