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
	Type        string   `yaml:"type"`
	Description string   `yaml:"description"`
	Optional    bool     `yaml:"optional"`
	Paths       []string `yaml:"paths"`
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
