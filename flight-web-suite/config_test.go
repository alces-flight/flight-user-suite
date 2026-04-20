package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigUsesEmbeddedDefaultsWhenFileMissing(t *testing.T) {
	root := t.TempDir()
	setFlightRootForTest(t, root)

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	defaultConfig := defaultConfig()

	if config.Port != defaultConfig.Port {
		t.Fatalf("expected default port %d, got %d", defaultConfig.Port, config.Port)
	}
	if config.Session.Secret == defaultConfig.Session.Secret {
		t.Fatalf("expected session secret to have changed from default %q, got %q", defaultConfig.Session.Secret, config.Session.Secret)
	}
	if config.Authenticator.Timeout != defaultConfig.Authenticator.Timeout {
		t.Fatalf("expected default authenticator timeout %s, got %s", defaultConfig.Authenticator.Timeout, config.Authenticator.Timeout)
	}
	if config.Authenticator.PAMService != defaultConfig.Authenticator.PAMService {
		t.Fatalf("expected default PAM service %q, got %q", defaultConfig.Authenticator.PAMService, config.Authenticator.PAMService)
	}
}

func TestLoadConfigUsesValuesFromFile(t *testing.T) {
	root := t.TempDir()
	setFlightRootForTest(t, root)

	configPath := filepath.Join(root, "etc", "web-suite.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}
	data := []byte(`port: 9090
session:
  secret: "configured-secret"
authenticator:
  timeout: "250ms"
  pam_service: "sshd"
`)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}

	if config.Port != 9090 {
		t.Fatalf("expected configured port 9090, got %d", config.Port)
	}
	if config.Session.Secret != "configured-secret" {
		t.Fatalf("expected configured session secret, got %q", config.Session.Secret)
	}
	if config.Authenticator.Timeout != 250*time.Millisecond {
		t.Fatalf("expected configured timeout 250ms, got %s", config.Authenticator.Timeout)
	}
	if config.Authenticator.PAMService != "sshd" {
		t.Fatalf("expected configured PAM service %q, got %q", "sshd", config.Authenticator.PAMService)
	}
}

func TestLoadConfigUsesDefaultValuesIfOmittedFromFile(t *testing.T) {
	root := t.TempDir()
	setFlightRootForTest(t, root)

	configPath := filepath.Join(root, "etc", "web-suite.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}
	data := []byte(`
session:
  secret: "configured-secret"
authenticator:
  timeout: "250ms"
`)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	defaultConfig := defaultConfig()

	if config.Port != defaultConfig.Port {
		t.Fatalf("expected default port %d, got %d", defaultConfig.Port, config.Port)
	}
	if config.Session.Secret != "configured-secret" {
		t.Fatalf("expected configured session secret, got %q", config.Session.Secret)
	}
	if config.Authenticator.Timeout != 250*time.Millisecond {
		t.Fatalf("expected configured timeout 250ms, got %s", config.Authenticator.Timeout)
	}
	if config.Authenticator.PAMService != defaultConfig.Authenticator.PAMService {
		t.Fatalf("expected default PAM service %q, got %q", defaultConfig.Authenticator.PAMService, config.Authenticator.PAMService)
	}
}

func setFlightRootForTest(t *testing.T, root string) {
	t.Helper()

	orig := flightRoot
	flightRoot = root
	authenticatorPath = filepath.Join(flightRoot, "usr", "libexec", "web-suite", "authenticate.py")
	t.Cleanup(func() {
		flightRoot = orig
		authenticatorPath = filepath.Join(flightRoot, "usr", "libexec", "web-suite", "authenticate.py")
	})
}
