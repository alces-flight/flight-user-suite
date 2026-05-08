package desktop

import (
	"log/slog"
	"path/filepath"
	"time"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type DesktopCli struct {
	Config RemoteConfig
	Env    configenv.Env
	Logger *slog.Logger
}

func (cli *DesktopCli) ToolPath() string {
	return filepath.Join(cli.Env.FlightRoot, "usr", "lib", "flight-core", "flight-desktop")
}

type RemoteConfig struct {
	SshKeyName        string        `yaml:"ssh_key_name"`
	HomeDirFallback   string        `yaml:"home_dir_fallback"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
}
