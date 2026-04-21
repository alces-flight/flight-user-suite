package main

import (
	"crypto/rand"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type webSuiteConfig struct {
	Port          int                 `yaml:"port"`
	Session       sessionConfig       `yaml:"-"`
	Authenticator authenticatorConfig `yaml:"authenticator"`
}

type sessionConfig struct {
	Secret string `yaml:"-"`
}

type authenticatorConfig struct {
	Timeout    time.Duration `yaml:"timeout"`
	PAMService string        `yaml:"pam_service"`
}

//go:embed opt/flight/etc/web-suite.yml
var _defaultConfig []byte

func defaultConfig() webSuiteConfig {
	var dc webSuiteConfig
	if err := yaml.Unmarshal(_defaultConfig, &dc); err != nil {
		panic(fmt.Errorf("loading default config: %w", err))
	}
	return dc
}

func loadConfig() (webSuiteConfig, error) {
	configPath := filepath.Join(flightRoot, "etc", "web-suite.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if pathError, ok := errors.AsType[*fs.PathError](err); ok && pathError.Err.Error() == "no such file or directory" {
			data = _defaultConfig
		} else {
			return webSuiteConfig{}, fmt.Errorf("loading config: %w", err)
		}
	}

	var config webSuiteConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return webSuiteConfig{}, fmt.Errorf("loading config from %s: %w", configPath, err)
	}

	config.Authenticator.PAMService = strings.TrimSpace(config.Authenticator.PAMService)

	// Merge defaults
	defaultConfig := defaultConfig()
	if config.Port == 0 {
		config.Port = defaultConfig.Port
	}
	if config.Authenticator.PAMService == "" {
		config.Authenticator.PAMService = defaultConfig.Authenticator.PAMService
	}

	if config.Authenticator.Timeout == 0 {
		config.Authenticator.Timeout = defaultConfig.Authenticator.Timeout
	}
	config.Session.Secret, err = loadSessionSecret()
	if err != nil {
		return webSuiteConfig{}, err
	}

	return config, nil
}

func loadSessionSecret() (string, error) {
	path := filepath.Join(flightStateRoot, "web-suite", "session-secret")
	data, err := os.ReadFile(path)

	// The file can't be read, for some reason other than it not existing. This
	// is a genuine error that should be returned.
	if err != nil {
		if pathError, ok := errors.AsType[*fs.PathError](err); !ok || pathError.Err.Error() != "no such file or directory" {
			return "", fmt.Errorf("loading session secret: %w", err)
		}
	}

	secret := strings.TrimSpace(string(data))
	if secret != "" {
		return secret, nil
	}

	secret = rand.Text()
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return "", fmt.Errorf("creating session directory: %w", err)
	}
	log.Printf("saving session secret path=%s\n", path)
	if err := os.WriteFile(path, []byte(secret), 0o600); err != nil {
		return "", fmt.Errorf("saving session secret: %w", err)
	}
	return secret, nil
}
