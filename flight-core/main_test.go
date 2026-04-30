package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/concertim/flight-user-suite/flight/configenv"
	"github.com/urfave/cli/v3"
)

func TestAddToolProxyCommandsSkipsDisabledTools(t *testing.T) {
	flightRoot := t.TempDir()
	toolPath := filepath.Join(flightRoot, "usr", "lib", "flight-core")
	synopsisPath := filepath.Join(flightRoot, "usr", "share", "doc", "tools")
	if err := os.MkdirAll(toolPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(synopsisPath, 0o755); err != nil {
		t.Fatal(err)
	}

	writeToolFixture(t, filepath.Join(toolPath, "flight-enabled"), 0o755, "enabled")
	writeToolFixture(t, filepath.Join(toolPath, "flight-disabled"), 0o644, "disabled")
	writeToolFixture(t, filepath.Join(synopsisPath, "flight-enabled"), 0o644, "Enabled synopsis")
	writeToolFixture(t, filepath.Join(synopsisPath, "flight-disabled"), 0o644, "Disabled synopsis")

	prevEnv := env
	t.Cleanup(func() {
		env = prevEnv
	})
	env = configenv.Env{FlightRoot: flightRoot}

	cmd := &cli.Command{}
	addToolProxyCommands(cmd)

	if len(cmd.Commands) != 1 {
		t.Fatalf("expected 1 proxy command, got %d", len(cmd.Commands))
	}
	if cmd.Commands[0].Name != "enabled" {
		t.Fatalf("expected only enabled tool to be exposed, got %q", cmd.Commands[0].Name)
	}
}

func writeToolFixture(t *testing.T, path string, mode os.FileMode, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}
