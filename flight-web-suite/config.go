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
	Session       sessionConfig       `yaml:"session"`
	Authenticator authenticatorConfig `yaml:"authenticator"`
}

type sessionConfig struct {
	Secret string `yaml:"secret"`
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
	path := filepath.Join(flightRoot, "etc", "web-suite.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		if pathError, ok := errors.AsType[*fs.PathError](err); ok && pathError.Err.Error() == "no such file or directory" {
			data = _defaultConfig
		} else {
			return webSuiteConfig{}, fmt.Errorf("loading config: %w", err)
		}
	}

	var config webSuiteConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return webSuiteConfig{}, fmt.Errorf("loading config from %s: %w", path, err)
	}

	config.Session.Secret = strings.TrimSpace(config.Session.Secret)
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
	// NOTE: Check session secret last as we may write the config file. If we
	// write too soon, we risk saving zero values into the config file, better
	// instead to save the default values.
	if config.Session.Secret == "" || config.Session.Secret == defaultConfig.Session.Secret {
		secret := rand.Text()
		config.Session.Secret = secret
		err := writeConfig(config, path)
		if err != nil {
			return webSuiteConfig{}, fmt.Errorf("saving session secret: %w", err)
		}
	}

	return config, nil
}

func writeConfig(config webSuiteConfig, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	log.Printf("saving session secret path=%s\n", path)
	parent := filepath.Dir(path)
	err = os.MkdirAll(parent, 0o700)
	if err != nil {
		return fmt.Errorf("creating session directory: %w", err)
	}
	return os.WriteFile(path, data, 0o600)
}
