package main

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/concertim/flight-user-suite/flight-web-suite/internal/testutil"
)

func TestNewDesktopSessionRedirectsForAnonymous(t *testing.T) {
	resp, _ := testutil.RenderPage(t, newApp(), http.MethodGet, "/desktop/new", nil, http.StatusSeeOther)

	if resp.Header.Get("location") != "/sessions" {
		t.Errorf("expected desktop launch page to redirect to '/sessions' for anonymous users")
	}
}

func TestCreateDesktopSessionRedirectsForAnonymous(t *testing.T) {
	resp, _ := testutil.RenderPage(
		t,
		newApp(),
		http.MethodPost,
		"/desktop",
		strings.NewReader(url.Values{"desktop_type": {"xterm"}}.Encode()),
		http.StatusSeeOther,
		testutil.WithContentType("application/x-www-form-urlencoded"),
	)

	if resp.Header.Get("location") != "/sessions" {
		t.Errorf("expected desktop launch to redirect to '/sessions' for anonymous users")
	}
}

func TestNewDesktopSessionReturnsServiceUnavailableWhenToolDisabled(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o644, "#!/bin/sh\n"))

	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/desktop/new", nil, http.StatusServiceUnavailable, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	if want := "Flight Desktop is not enabled"; !strings.Contains(body, want) {
		t.Fatalf("expected body to contain %q, got %q", want, body)
	}
}

func TestCreateDesktopSessionReturnsServiceUnavailableWhenToolDisabled(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o644, "#!/bin/sh\n"))

	_, body := testutil.RenderPage(
		t,
		newApp(),
		http.MethodPost,
		"/desktop",
		strings.NewReader(url.Values{"desktop_type": {"xterm"}}.Encode()),
		http.StatusServiceUnavailable,
		testutil.WithContentType("application/x-www-form-urlencoded"),
		testutil.WithSessionCookie(currentUser.Username, config.Session.Secret),
	)

	if want := "Flight Desktop is not enabled"; !strings.Contains(body, want) {
		t.Fatalf("expected body to contain %q, got %q", want, body)
	}
}

func TestNewDesktopSessionPageDisplaysAvailableTypesOrderedByIDAndSelectsFirst(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
if [ "$1" = "avail" ]; then
cat <<'JSON'
[
  {
    "id": "xterm",
    "summary": "Minimal terminal desktop.",
    "url": "https://example.invalid/xterm"
  },
  {
    "id": "gnome",
    "summary": "GNOME desktop.",
    "url": "https://example.invalid/gnome"
  }
]
JSON
fi
`))

	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/desktop/new", nil, http.StatusOK, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	testutil.AssertSelection(t, body, `[data-testid="desktop-launch-back-link"]`, testutil.HasAttr("href", "/desktop"))
	testutil.AssertSelection(t, body, `[data-testid="desktop-type-card--gnome"] h3`, testutil.HasText("gnome"))
	testutil.AssertSelection(t, body, `[data-testid="desktop-type-card--gnome"] h3 + p`, testutil.HasText("GNOME desktop."))
	testutil.AssertSelection(t, body, `[data-testid="desktop-type-card--gnome"] a`, testutil.HasAttr("href", "https://example.invalid/gnome"))
	testutil.AssertSelection(t, body, `[data-testid="desktop-type-input--gnome"]`, testutil.HasAttr("checked", ""))
	testutil.AssertSelection(t, body, `[data-testid="desktop-launch-geometry-option--1024x768"]`, testutil.HasAttr("selected", ""))

	gnomeIndex := strings.Index(body, `data-testid="desktop-type-card--gnome"`)
	xtermIndex := strings.Index(body, `data-testid="desktop-type-card--xterm"`)
	if gnomeIndex == -1 || xtermIndex == -1 || gnomeIndex > xtermIndex {
		t.Fatalf("expected desktop type cards to be ordered by id, got body:\n%s", body)
	}
}

func TestDesktopSessionsPageLinksToDesktopLaunchPage(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, "#!/bin/sh\necho '[]'\n"))

	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/desktop", nil, http.StatusOK, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	testutil.AssertSelection(t, body, `[data-testid="launch-desktop-link"]`, testutil.HasAttr("href", "/desktop/new"), testutil.HasText("Launch Desktop"))
}

func TestNewDesktopSessionInvokesAvailWithJSONFormat(t *testing.T) {
	currentUser := currentUserForTest(t)
	argsFile := filepath.Join(t.TempDir(), "desktop-args.txt")
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
printf '%s\n' "$@" > "`+argsFile+`"
echo '[]'
`))

	testutil.RenderPage(t, newApp(), http.MethodGet, "/desktop/new", nil, http.StatusOK, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read command args fixture: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "avail\n--format\njson" {
		t.Fatalf("expected avail command args, got %q", got)
	}
}

func TestCreateDesktopSessionInvokesStartWithJSONFormatAndOmitsEmptyName(t *testing.T) {
	currentUser := currentUserForTest(t)
	argsFile := filepath.Join(t.TempDir(), "desktop-args.txt")
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
printf '%s\n' "$@" > "`+argsFile+`"
if [ "$1" = "avail" ]; then
cat <<'JSON'
[
  {
    "id": "xterm",
    "summary": "Minimal terminal desktop.",
    "url": "https://example.invalid/xterm"
  }
]
JSON
exit 0
fi
cat <<'JSON'
{
  "success": true,
  "session_name": "xterm.abc123"
}
JSON
`))

	resp, _ := testutil.RenderPage(
		t,
		newApp(),
		http.MethodPost,
		"/desktop",
		strings.NewReader(url.Values{
			"desktop_type": {"xterm"},
			"name":         {""},
			"geometry":     {"1024x768"},
		}.Encode()),
		http.StatusSeeOther,
		testutil.WithContentType("application/x-www-form-urlencoded"),
		testutil.WithSessionCookie(currentUser.Username, config.Session.Secret),
	)

	if resp.Header.Get("location") != "/desktop" {
		t.Fatalf("expected redirect to /desktop, got %q", resp.Header.Get("location"))
	}

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read command args fixture: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "start\nxterm\n--format\njson\n--geometry\n1024x768" {
		t.Fatalf("expected start command args without name, got %q", got)
	}
}

func TestCreateDesktopSessionInvokesStartWithProvidedName(t *testing.T) {
	currentUser := currentUserForTest(t)
	argsFile := filepath.Join(t.TempDir(), "desktop-args.txt")
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
printf '%s\n' "$@" > "`+argsFile+`"
if [ "$1" = "avail" ]; then
cat <<'JSON'
[
  {
    "id": "xterm",
    "summary": "Minimal terminal desktop.",
    "url": "https://example.invalid/xterm"
  }
]
JSON
exit 0
fi
cat <<'JSON'
{
  "success": true,
  "session_name": "named-session"
}
JSON
`))

	testutil.RenderPage(
		t,
		newApp(),
		http.MethodPost,
		"/desktop",
		strings.NewReader(url.Values{
			"desktop_type": {"xterm"},
			"name":         {"named-session"},
			"geometry":     {"1280x1024"},
		}.Encode()),
		http.StatusSeeOther,
		testutil.WithContentType("application/x-www-form-urlencoded"),
		testutil.WithSessionCookie(currentUser.Username, config.Session.Secret),
	)

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read command args fixture: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "start\nxterm\n--format\njson\n--geometry\n1280x1024\n--name\nnamed-session" {
		t.Fatalf("expected start command args with name, got %q", got)
	}
}

func TestCreateDesktopSessionFailureRedisplaysFormWithPreservedValues(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
if [ "$1" = "avail" ]; then
cat <<'JSON'
[
  {
    "id": "gnome",
    "summary": "GNOME desktop.",
    "url": "https://example.invalid/gnome"
  },
  {
    "id": "xterm",
    "summary": "Minimal terminal desktop.",
    "url": "https://example.invalid/xterm"
  }
]
JSON
exit 0
fi
cat <<'JSON'
{
  "errors": [
    {
      "code": "invalid_name",
      "title": "Invalid session name",
      "detail": "Invalid session name.",
      "source": {
        "parameter": "name"
      }
    }
  ]
}
JSON
exit 1
`))

	_, body := testutil.RenderPage(
		t,
		newApp(),
		http.MethodPost,
		"/desktop",
		strings.NewReader(url.Values{
			"desktop_type": {"xterm"},
			"name":         {"custom-session"},
			"geometry":     {"1280x1024"},
		}.Encode()),
		http.StatusUnprocessableEntity,
		testutil.WithContentType("application/x-www-form-urlencoded"),
		testutil.WithSessionCookie(currentUser.Username, config.Session.Secret),
	)

	testutil.AssertSelection(t, body, `div.flash.alert`, testutil.HasText("Failed to launch desktop session."))
	testutil.AssertSelection(t, body, `[data-testid="desktop-launch-name"]`, testutil.HasAttr("value", "custom-session"))
	testutil.AssertSelection(t, body, `[data-testid="desktop-type-input--xterm"]`, testutil.HasAttr("checked", ""))
	testutil.AssertSelection(t, body, `[data-testid="desktop-launch-geometry-option--1280x1024"]`, testutil.HasAttr("selected", ""))
	testutil.AssertSelection(t, body, `[data-testid="desktop-launch-name-error"]`, testutil.HasText("Invalid session name."))
}

func TestCreateDesktopSessionFailureWithoutProvidedNameDoesNotExposeGeneratedName(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
if [ "$1" = "avail" ]; then
cat <<'JSON'
[
  {
    "id": "xterm",
    "summary": "Minimal terminal desktop.",
    "url": "https://example.invalid/xterm"
  }
]
JSON
exit 0
fi
cat <<'JSON'
{
  "errors": [
    {
      "code": "start_failed",
      "title": "Desktop session failed to start",
      "detail": "Starting desktop session failed."
    }
  ]
}
JSON
exit 1
`))

	_, body := testutil.RenderPage(
		t,
		newApp(),
		http.MethodPost,
		"/desktop",
		strings.NewReader(url.Values{
			"desktop_type": {"xterm"},
			"name":         {""},
			"geometry":     {"1024x768"},
		}.Encode()),
		http.StatusUnprocessableEntity,
		testutil.WithContentType("application/x-www-form-urlencoded"),
		testutil.WithSessionCookie(currentUser.Username, config.Session.Secret),
	)

	testutil.AssertSelection(t, body, `div.flash.alert`, testutil.HasText("Failed to launch desktop session."))
	if strings.Contains(body, "xterm.secret-name") {
		t.Fatalf("expected generated session name to be hidden on launch failure, got body:\n%s", body)
	}
}

func TestCreateDesktopSessionInvalidNameShowsUserFriendlyMessage(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
if [ "$1" = "avail" ]; then
cat <<'JSON'
[
  {
    "id": "xterm",
    "summary": "Minimal terminal desktop.",
    "url": "https://example.invalid/xterm"
  }
]
JSON
exit 0
fi
cat <<'JSON'
{
  "errors": [
    {
      "code": "invalid_name",
      "title": "Invalid session name",
      "detail": "Session name can contain only letters, numbers, hyphens, underscores and dots and cannot start with a hyphen.",
      "source": {
        "parameter": "name"
      }
    }
  ]
}
JSON
exit 1
`))

	_, body := testutil.RenderPage(
		t,
		newApp(),
		http.MethodPost,
		"/desktop",
		strings.NewReader(url.Values{
			"desktop_type": {"xterm"},
			"name":         {"--foo"},
			"geometry":     {"1024x768"},
		}.Encode()),
		http.StatusUnprocessableEntity,
		testutil.WithContentType("application/x-www-form-urlencoded"),
		testutil.WithSessionCookie(currentUser.Username, config.Session.Secret),
	)

	want := "Session name can contain only letters, numbers, hyphens, underscores and dots and cannot start with a hyphen."
	testutil.AssertSelection(t, body, `[data-testid="desktop-launch-name-error"]`, testutil.HasText(want))
	if strings.Contains(body, "for flag -name") {
		t.Fatalf("expected user-facing validation message without CLI flag details, got body:\n%s", body)
	}
}

func TestCreateDesktopSessionWithInvalidGeometryShowsInlineErrorWithoutInvokingStart(t *testing.T) {
	currentUser := currentUserForTest(t)
	argsFile := filepath.Join(t.TempDir(), "desktop-args.txt")
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
printf '%s\n' "$@" >> "`+argsFile+`"
cat <<'JSON'
[
  {
    "id": "xterm",
    "summary": "Minimal terminal desktop.",
    "url": "https://example.invalid/xterm"
  }
]
JSON
`))

	_, body := testutil.RenderPage(
		t,
		newApp(),
		http.MethodPost,
		"/desktop",
		strings.NewReader(url.Values{
			"desktop_type": {"xterm"},
			"name":         {"bad-geometry"},
			"geometry":     {"640x480"},
		}.Encode()),
		http.StatusUnprocessableEntity,
		testutil.WithContentType("application/x-www-form-urlencoded"),
		testutil.WithSessionCookie(currentUser.Username, config.Session.Secret),
	)

	testutil.AssertSelection(t, body, `div.flash.alert`, testutil.HasText("Failed to launch desktop session."))
	testutil.AssertSelection(t, body, `[data-testid="desktop-launch-geometry-error"]`, testutil.HasText("Select a supported geometry."))
	testutil.AssertSelection(t, body, `[data-testid="desktop-launch-name"]`, testutil.HasAttr("value", "bad-geometry"))

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read command args fixture: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "avail\n--format\njson" {
		t.Fatalf("expected only avail command args, got %q", got)
	}
}

func TestCreateDesktopSessionWithInvalidDesktopTypeShowsInlineErrorWithoutInvokingStart(t *testing.T) {
	currentUser := currentUserForTest(t)
	argsFile := filepath.Join(t.TempDir(), "desktop-args.txt")
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
printf '%s\n' "$@" >> "`+argsFile+`"
cat <<'JSON'
[
  {
    "id": "xterm",
    "summary": "Minimal terminal desktop.",
    "url": "https://example.invalid/xterm"
  }
]
JSON
`))

	_, body := testutil.RenderPage(
		t,
		newApp(),
		http.MethodPost,
		"/desktop",
		strings.NewReader(url.Values{
			"desktop_type": {"gnome"},
			"name":         {"bad-type"},
			"geometry":     {"1024x768"},
		}.Encode()),
		http.StatusUnprocessableEntity,
		testutil.WithContentType("application/x-www-form-urlencoded"),
		testutil.WithSessionCookie(currentUser.Username, config.Session.Secret),
	)

	testutil.AssertSelection(t, body, `div.flash.alert`, testutil.HasText("Failed to launch desktop session."))
	testutil.AssertSelection(t, body, `[data-testid="desktop-launch-desktop-type-error"]`, testutil.HasText("Select an available desktop type."))
	testutil.AssertSelection(t, body, `[data-testid="desktop-launch-name"]`, testutil.HasAttr("value", "bad-type"))
	testutil.AssertSelection(t, body, `[data-testid="desktop-launch-geometry-option--1024x768"]`, testutil.HasAttr("selected", ""))

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read command args fixture: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "avail\n--format\njson" {
		t.Fatalf("expected only avail command args, got %q", got)
	}
}
