package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type config struct {
	Dependencies []dependency `yaml:"dependencies"`
}

type dependency struct {
	Type  string   `yaml:"type"`
	Paths []string `yaml:"paths"`
}

func loadConfig() (*config, error) {
	path := filepath.Join(flightRoot, "etc", "desktop.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	var config config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("loading config from %s: %w", path, err)
	}
	return &config, nil
}
