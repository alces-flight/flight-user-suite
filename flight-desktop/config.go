package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/concertim/flight-user-suite/flight/doctor"
	"gopkg.in/yaml.v3"
)

type desktopConfig struct {
	NameGenerator nameGeneratorConfig `yaml:"name_generator"`
	Dependencies  []doctor.Dependency `yaml:"dependencies"`
	EnvWhitelist  []string            `yaml:"environment_whitelist"`
	VncPasswd     string              `yaml:"vncpasswd"`
	WebSockify    string              `yaml:"websockify"`
	ScreenGrabber string              `yaml:"screen_grabber"`
}

type nameGeneratorConfig struct {
	Strategy string `yaml:"strategy"`
}

//go:embed opt/flight/etc/desktop.yml
var defaultConfig []byte

func loadConfig() (desktopConfig, error) {
	path := filepath.Join(env.FlightRoot, "etc", "desktop.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		if pathError, ok := errors.AsType[*fs.PathError](err); ok && pathError.Err.Error() == "no such file or directory" {
			data = defaultConfig
		} else {
			return desktopConfig{}, fmt.Errorf("loading config: %w", err)
		}
	}
	var config desktopConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return desktopConfig{}, fmt.Errorf("loading config from %s: %w", path, err)
	}

	if config.NameGenerator.Strategy == "" {
		config.NameGenerator.Strategy = "absurd"
	}
	if len(config.EnvWhitelist) == 0 {
		config.EnvWhitelist = []string{"PWD", "HOME", "LANG", "USER", "UID", "PATH", "VNCDESKTOP", "DISPLAY", "FLIGHT_ROOT"}
	} else {
		whitelist := make([]string, 0, len(config.EnvWhitelist))
		for _, item := range config.EnvWhitelist {
			whitelist = append(whitelist, strings.TrimSpace(item))
		}
		config.EnvWhitelist = whitelist
	}
	if config.VncPasswd == "" {
		config.VncPasswd = "/usr/bin/vncpasswd"
	}

	return config, nil
}
