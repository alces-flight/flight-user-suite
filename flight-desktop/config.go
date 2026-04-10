package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type desktopConfig struct {
	NameGenerator nameGeneratorConfig `yaml:"name_generator"`
	Dependencies  []dependency        `yaml:"dependencies"`
	EnvWhitelist  []string            `yaml:"environment_whitelist"`
	VncPasswd     string              `yaml:"vncpasswd"`
}

type nameGeneratorConfig struct {
	Strategy string `yaml:"strategy"`
}

type dependency struct {
	Type           string   `yaml:"type"`
	Description    string   `yaml:"description"`
	Optional       bool     `yaml:"optional"`
	Paths          []string `yaml:"paths"`
	FailureMessage string   `yaml:"failure_message"`
	SuccessMessage string   `yaml:"success_message"`
}

func requiredDependencies(deps []dependency) []dependency {
	opt := make([]dependency, 0)
	for _, dep := range deps {
		if !dep.Optional {
			opt = append(opt, dep)
		}
	}
	return opt
}

func optionalDependencies(deps []dependency) []dependency {
	opt := make([]dependency, 0)
	for _, dep := range deps {
		if dep.Optional {
			opt = append(opt, dep)
		}
	}
	return opt
}

//go:embed opt/flight/etc/desktop.yml
var defaultConfig []byte

func loadConfig() (desktopConfig, error) {
	path := filepath.Join(flightRoot, "etc", "desktop.yml")
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
