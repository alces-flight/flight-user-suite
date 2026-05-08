package cli

import (
	"log/slog"
	"time"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type Tool interface {
	ToolPath() string
	GetEnv() configenv.Env
	Logger() *slog.Logger
}

type RemoteTool interface {
	Tool
	RemoteConfig() RemoteConfig
}

type RemoteConfig struct {
	SshKeyName        string        `yaml:"ssh_key_name"`
	HomeDirFallback   string        `yaml:"home_dir_fallback"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
}
