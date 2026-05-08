package howto

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

func TestListCommandDecodesSuccessfulEnvelope(t *testing.T) {
	env := howtoEnvForTest(t, howtoToolFixture(t, 0o755, `#!/bin/sh
cat <<'JSON'
{
  "success": true,
  "guides": [
    {
      "index": 1,
      "title": "Base Guide A"
    }
  ]
}
JSON
`))

	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("failed to determine current user: %v", err)
	}

	response, err := newHowtoCliForTest(env).ListCommand(context.Background(), currentUser.Username)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !response.Success {
		t.Fatalf("expected success response, got %+v", response)
	}
	if len(response.Guides) != 1 || response.Guides[0].Index != 1 || response.Guides[0].Title != "Base Guide A" {
		t.Fatalf("unexpected guides: %+v", response.Guides)
	}
}

func TestListCommandDecodesFailureEnvelopeFromNonZeroExit(t *testing.T) {
	env := howtoEnvForTest(t, howtoToolFixture(t, 0o755, `#!/bin/sh
cat <<'JSON'
{
  "success": false,
  "guides": [],
  "error": "No such file or directory",
  "reason": "unexpected"
}
JSON
exit 1
`))

	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("failed to determine current user: %v", err)
	}

	response, err := newHowtoCliForTest(env).ListCommand(context.Background(), currentUser.Username)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.Success {
		t.Fatalf("expected failure response, got %+v", response)
	}
	if response.Error != "No such file or directory" || response.Reason != "unexpected" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestShowCommandDecodesSuccessfulEnvelope(t *testing.T) {
	env := howtoEnvForTest(t, howtoToolFixture(t, 0o755, `#!/bin/sh
cat <<'JSON'
{
  "success": true,
  "guide": {
    "title": "Base Guide A",
    "raw_markdown": "# Heading\n\nBody text\n"
  }
}
JSON
`))

	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("failed to determine current user: %v", err)
	}

	response, err := newHowtoCliForTest(env).ShowCommand(context.Background(), currentUser.Username, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !response.Success {
		t.Fatalf("expected success response, got %+v", response)
	}
	if response.Guide.Title != "Base Guide A" || response.Guide.RawMarkdown != "# Heading\n\nBody text\n" {
		t.Fatalf("unexpected guide: %+v", response.Guide)
	}
}

func TestShowCommandDecodesFailureEnvelopeFromNonZeroExit(t *testing.T) {
	env := howtoEnvForTest(t, howtoToolFixture(t, 0o755, `#!/bin/sh
cat <<'JSON'
{
  "success": false,
  "guide": {},
  "error": "Unknown howto: 01-base-guide-a.md",
  "reason": "not_found"
}
JSON
exit 1
`))

	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("failed to determine current user: %v", err)
	}

	response, err := newHowtoCliForTest(env).ShowCommand(context.Background(), currentUser.Username, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if response.Success {
		t.Fatalf("expected failure response, got %+v", response)
	}
	if response.Error != "Unknown howto: 01-base-guide-a.md" || response.Reason != "not_found" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func howtoEnvForTest(t *testing.T, root string) configenv.Env {
	t.Helper()

	env, err := configenv.InitFlightEnv()
	if err != nil {
		t.Fatalf("failed to init env: %v", err)
	}
	env.FlightRoot = root
	return env
}

func howtoToolFixture(t *testing.T, mode os.FileMode, script string) string {
	t.Helper()

	root := t.TempDir()
	toolPath := filepath.Join(root, "usr", "lib", "flight-core", "flight-howto")
	if err := os.MkdirAll(filepath.Dir(toolPath), 0o755); err != nil {
		t.Fatalf("failed to create tool dir: %v", err)
	}
	if err := os.WriteFile(toolPath, []byte(script), mode); err != nil {
		t.Fatalf("failed to write howto tool fixture: %v", err)
	}
	return root
}

func envContains(env []string, key, want string) bool {
	prefix := key + "="
	for _, entry := range env {
		if after, ok := strings.CutPrefix(entry, prefix); ok {
			return after == want
		}
	}
	return false
}

func newHowtoCliForTest(env configenv.Env) *HowtoCli {
	return &HowtoCli{Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Env: env}
}
