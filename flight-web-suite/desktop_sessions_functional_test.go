package main

import (
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"github.com/concertim/flight-user-suite/flight-web-suite/internal/testutil"
)

func TestDesktopSessionsRedirectsForAnonymous(t *testing.T) {
	resp, _ := testutil.RenderPage(t, newApp(), http.MethodGet, "/desktop", nil, http.StatusSeeOther)

	if resp.Header.Get("location") != "/sessions" {
		t.Errorf("expected desktop sessions page to redirect to '/sessions' for anonymous users")
	}
}

func TestDesktopSessionsPageDisplaysSessions(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
cat <<'JSON'
[
  {
    "name": "alpha",
    "desktop_type": "Gnome",
    "state": "active",
    "host": "host-a",
    "created_at": "2026-04-20T10:00:00Z"
  },
  {
    "name": "beta",
    "desktop_type": "Xfce",
    "state": "broken",
    "host": "host-b",
    "created_at": "2026-04-21T11:30:00Z"
  }
]
JSON
`))

	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/desktop", nil, http.StatusOK, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	assertAuthenticated(t, body, currentUser.Username)
	testutil.AssertSelection(t, body, `[data-testid="desktop-sessions-summary"]`,
		testutil.HasText("You have 1 desktop session currently running and 1 stale session."),
	)
	testutil.AssertSelection(t, body, `[data-testid="desktop-session-card--alpha"] p`,
		testutil.HasText("alpha"),
	)
	testutil.AssertSelection(t, body, `[data-testid="desktop-session-desktop-type--alpha"]`,
		testutil.HasText("Gnome"),
	)
	testutil.AssertSelection(t, body, `[data-testid="desktop-session-state--alpha"]`,
		testutil.HasText("active"),
	)
	testutil.AssertSelection(t, body, `[data-testid="desktop-session-host--alpha"]`,
		testutil.HasText("host-a"),
	)
	testutil.AssertSelection(t, body, `[data-testid="desktop-session-start-time--alpha"]`,
		testutil.HasText("Mon 20 Apr 2026 10:00"),
	)
	testutil.AssertSelection(t, body, `[data-testid="desktop-session-action-button--alpha"]`,
		testutil.HasText("Terminate"),
	)
	testutil.AssertSelection(t, body, `form[action="/desktop/alpha"] input[name="_method"]`,
		testutil.HasAttr("value", "DELETE"),
	)
	testutil.AssertSelection(t, body, `[data-testid="desktop-session-card--beta"] p`,
		testutil.HasText("beta"),
	)
	testutil.AssertSelection(t, body, `form[action="/desktop/beta/clean"] [data-testid="desktop-session-action-button--beta"]`)
	testutil.AssertSelection(t, body, `[data-testid="desktop-session-action-button--beta"]`,
		testutil.HasText("Remove"),
	)
}

func TestDesktopSessionsPageReturnsServiceUnavailableWhenToolDisabled(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o644, `#!/bin/sh
echo '[]'
`))

	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/desktop", nil, http.StatusServiceUnavailable, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	if want := "Flight Desktop is not enabled"; !strings.Contains(body, want) {
		t.Fatalf("expected body to contain %q, got %q", want, body)
	}
}

func TestDesktopSessionsPageShowsRemoteActionAsDisabledTerminate(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
cat <<'JSON'
[
  {
    "name": "remote-a",
    "desktop_type": "Gnome",
    "state": "remote",
    "host": "remote-host",
    "created_at": "2026-04-20T10:00:00Z"
  }
]
JSON
`))

	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/desktop", nil, http.StatusOK, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	testutil.AssertSelection(t, body, `[data-testid="desktop-session-action-button--remote-a"]`,
		testutil.HasText("Terminate"),
		testutil.HasAttr("disabled", ""),
		testutil.HasAttr("title", "Termination of remote sessions is not yet implemented."),
	)
}

func TestDestroyDesktopSessionRedirectsForAnonymous(t *testing.T) {
	resp, _ := testutil.RenderPage(t, newApp(), http.MethodDelete, "/desktop/alpha", nil, http.StatusSeeOther)

	if resp.Header.Get("location") != "/sessions" {
		t.Errorf("expected desktop termination to redirect to '/sessions' for anonymous users")
	}
}

func TestDestroyDesktopSessionReturnsServiceUnavailableWhenToolDisabled(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o644, `#!/bin/sh
echo '[]'
`))

	_, body := testutil.RenderPage(t, newApp(), http.MethodDelete, "/desktop/alpha", nil, http.StatusServiceUnavailable, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	if want := "Flight Desktop is not enabled"; !strings.Contains(body, want) {
		t.Fatalf("expected body to contain %q, got %q", want, body)
	}
}

func TestCleanDesktopSessionRedirectsForAnonymous(t *testing.T) {
	resp, _ := testutil.RenderPage(t, newApp(), http.MethodPost, "/desktop/alpha/clean", nil, http.StatusSeeOther)

	if resp.Header.Get("location") != "/sessions" {
		t.Errorf("expected desktop clean to redirect to '/sessions' for anonymous users")
	}
}

func TestCleanDesktopSessionReturnsServiceUnavailableWhenToolDisabled(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o644, `#!/bin/sh
echo '[]'
`))

	_, body := testutil.RenderPage(t, newApp(), http.MethodPost, "/desktop/alpha/clean", nil, http.StatusServiceUnavailable, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	if want := "Flight Desktop is not enabled"; !strings.Contains(body, want) {
		t.Fatalf("expected body to contain %q, got %q", want, body)
	}
}

func TestDestroyDesktopSessionInvokesKillWithJSONFormat(t *testing.T) {
	currentUser := currentUserForTest(t)
	argsFile := filepath.Join(t.TempDir(), "desktop-args.txt")
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
printf '%s\n' "$@" > "`+argsFile+`"
cat <<'JSON'
{
  "success": true,
  "session_name": "alpha"
}
JSON
`))

	testutil.RenderPage(t, newApp(), http.MethodDelete, "/desktop/alpha", nil, http.StatusSeeOther, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read command args fixture: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "kill\n--format\njson\n--\nalpha" {
		t.Fatalf("expected kill command args, got %q", got)
	}
}

func TestCleanDesktopSessionInvokesCleanWithJSONFormat(t *testing.T) {
	currentUser := currentUserForTest(t)
	argsFile := filepath.Join(t.TempDir(), "desktop-args.txt")
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
printf '%s\n' "$@" > "`+argsFile+`"
cat <<'JSON'
{
  "success": true,
  "results": [
    {
      "success": true,
      "session_name": "alpha"
    }
  ]
}
JSON
`))

	testutil.RenderPage(t, newApp(), http.MethodPost, "/desktop/alpha/clean", nil, http.StatusSeeOther, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read command args fixture: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "clean\n--format\njson\n--\nalpha" {
		t.Fatalf("expected clean command args, got %q", got)
	}
}

func currentUserForTest(t *testing.T) *user.User {
	t.Helper()

	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("failed to determine current user: %v", err)
	}
	return currentUser
}

func setFlightRootForDesktopTest(t *testing.T, root string) {
	t.Helper()

	original := env.FlightRoot
	env.FlightRoot = root
	t.Cleanup(func() {
		env.FlightRoot = original
	})
}

func desktopToolFixture(t *testing.T, mode os.FileMode, script string) string {
	t.Helper()

	root := t.TempDir()
	toolPath := filepath.Join(root, "usr", "lib", "flight-core", "flight-desktop")
	if err := os.MkdirAll(filepath.Dir(toolPath), 0o755); err != nil {
		t.Fatalf("failed to create tool dir: %v", err)
	}
	if err := os.WriteFile(toolPath, []byte(script), mode); err != nil {
		t.Fatalf("failed to write desktop tool fixture: %v", err)
	}
	synopsisDir := filepath.Join(root, "usr", "share", "doc", "tools")
	if err := os.MkdirAll(synopsisDir, 0o755); err != nil {
		t.Fatalf("failed to create synopsis dir: %v", err)
	}
	synopsisPath := filepath.Join(synopsisDir, "flight-desktop")
	if err := os.WriteFile(synopsisPath, []byte("Access interactive desktop sessions"), 0o644); err != nil {
		t.Fatalf("failed to write synopsis fixture: %v", err)
	}
	return root
}
