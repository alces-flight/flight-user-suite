// Package desktop provides functions and types for running the
// `flight-desktop` CLI and parsing its output.
package desktop

import (
	"log/slog"
	"path/filepath"

	"github.com/concertim/flight-user-suite/flight-web-suite/cli"
	"github.com/concertim/flight-user-suite/flight/configenv"
)

type DesktopCli struct {
	config cli.RemoteConfig
	env    configenv.Env
	logger *slog.Logger
}

func NewCliTool(logger *slog.Logger, env configenv.Env, config cli.RemoteConfig) *DesktopCli {
	return &DesktopCli{
		config: config,
		env:    env,
		logger: logger,
	}
}

func (dcli *DesktopCli) ToolPath() string {
	return filepath.Join(dcli.env.FlightRoot, "usr", "lib", "flight-core", "flight-desktop")
}

func (dcli *DesktopCli) GetEnv() configenv.Env {
	return dcli.env
}

func (dcli *DesktopCli) Logger() *slog.Logger {
	return dcli.logger
}

func (dcli *DesktopCli) RemoteConfig() cli.RemoteConfig {
	return dcli.config
}
