package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

func TestLoadConfigUsesEmbeddedDefaultsWhenFileMissing(t *testing.T) {
	root := t.TempDir()
	setFlightRootForTest(t, root)

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	defaultConfig := defaultConfig()

	assertGeneratedSessionSecret(t, config.Session.Secret)
	assertSessionSecretSaved(t, config.Session.Secret)
	if config.Port != defaultConfig.Port {
		t.Fatalf("expected default port %d, got %d", defaultConfig.Port, config.Port)
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

	assertGeneratedSessionSecret(t, config.Session.Secret)
	assertSessionSecretSaved(t, config.Session.Secret)
	if config.Port != 9090 {
		t.Fatalf("expected configured port 9090, got %d", config.Port)
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

	assertGeneratedSessionSecret(t, config.Session.Secret)
	assertSessionSecretSaved(t, config.Session.Secret)
	if config.Port != defaultConfig.Port {
		t.Fatalf("expected default port %d, got %d", defaultConfig.Port, config.Port)
	}
	if config.Authenticator.Timeout != 250*time.Millisecond {
		t.Fatalf("expected configured timeout 250ms, got %s", config.Authenticator.Timeout)
	}
	if config.Authenticator.PAMService != defaultConfig.Authenticator.PAMService {
		t.Fatalf("expected default PAM service %q, got %q", defaultConfig.Authenticator.PAMService, config.Authenticator.PAMService)
	}
}

func TestLoadConfigIgnoresSessionSecretInConfigFile(t *testing.T) {
	root := t.TempDir()
	setFlightRootForTest(t, root)

	configPath := filepath.Join(root, "etc", "web-suite.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}
	data := []byte(`session:
  secret: "configured-secret"
`)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}

	if config.Session.Secret == "configured-secret" {
		t.Fatalf("expected config-file session secret to be ignored, got %q", config.Session.Secret)
	}
	assertGeneratedSessionSecret(t, config.Session.Secret)
	assertSessionSecretSaved(t, config.Session.Secret)
}

func setFlightRootForTest(t *testing.T, root string) {
	t.Helper()

	orig := env
	env = configenv.RepoLocalFlightEnv(root)
	authenticatorPath = filepath.Join(env.FlightRoot, "usr", "libexec", "web-suite", "authenticate.py")
	t.Cleanup(func() {
		env = orig
		authenticatorPath = filepath.Join(env.FlightRoot, "usr", "libexec", "web-suite", "authenticate.py")
	})
}

func assertSessionSecretSaved(t *testing.T, expected string) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(env.FlightStateRoot, "web-suite", "session-secret"))
	if err != nil {
		t.Fatalf("failed to read persisted session secret: %v", err)
	}
	if string(data) != expected {
		t.Fatalf("expected persisted session secret %q, got %q", expected, string(data))
	}
}

func assertGeneratedSessionSecret(t *testing.T, secret string) {
	t.Helper()

	if secret == "" {
		t.Fatal("expected generated session secret, got empty string")
	}
	if secret == "change-me" {
		t.Fatalf("expected generated session secret, got default %q", secret)
	}
}
