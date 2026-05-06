package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/concertim/flight-user-suite/flight-web-suite/internal/testutil"
)

func TestHowtoIndexRedirectsForAnonymous(t *testing.T) {
	resp, _ := testutil.RenderPage(t, newApp(), http.MethodGet, "/howto", nil, http.StatusSeeOther)

	if resp.Header.Get("location") != "/sessions" {
		t.Errorf("expected howto page to redirect to '/sessions' for anonymous users")
	}
}

func TestHowtoShowRedirectsForAnonymous(t *testing.T) {
	resp, _ := testutil.RenderPage(t, newApp(), http.MethodGet, "/howto/1", nil, http.StatusSeeOther)

	if resp.Header.Get("location") != "/sessions" {
		t.Errorf("expected howto guide page to redirect to '/sessions' for anonymous users")
	}
}

func TestHowtoIndexReturnsServiceUnavailableWhenToolDisabled(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForHowtoTest(t, howtoToolFixtureForWeb(t, 0o644, "#!/bin/sh\necho '{}'\n"))

	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/howto", nil, http.StatusServiceUnavailable, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	if want := "Flight Howto is not enabled"; !strings.Contains(body, want) {
		t.Fatalf("expected body to contain %q, got %q", want, body)
	}
}

func TestHowtoShowReturnsServiceUnavailableWhenToolDisabled(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForHowtoTest(t, howtoToolFixtureForWeb(t, 0o644, "#!/bin/sh\necho '{}'\n"))

	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/howto/1", nil, http.StatusServiceUnavailable, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	if want := "Flight Howto is not enabled"; !strings.Contains(body, want) {
		t.Fatalf("expected body to contain %q, got %q", want, body)
	}
}

func TestHowtoIndexRendersSidebarGuideListAndSelectionPrompt(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForHowtoTest(t, howtoToolFixtureForWeb(t, 0o755, `#!/bin/sh
cat <<'JSON'
{
  "success": true,
  "guides": [
    {"index": 1, "title": "Base Guide A"},
    {"index": 2, "title": "Cluster Basics"}
  ]
}
JSON
`))

	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/howto", nil, http.StatusOK, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	assertAuthenticated(t, body, currentUser.Username)
	testutil.AssertSelection(t, body, `[data-testid="howto-guide-link--1"]`,
		testutil.HasAttr("href", "/howto/1"),
		testutil.HasText("Base Guide A"),
		testutil.HasNoAttr("aria-current"),
	)
	testutil.AssertSelection(t, body, `[data-testid="howto-guide-link--2"]`,
		testutil.HasAttr("href", "/howto/2"),
		testutil.HasText("Cluster Basics"),
		testutil.HasNoAttr("aria-current"),
	)
	testutil.AssertSelection(t, body, `[data-testid="howto-welcome"]`,
		testutil.HasText("Welcome to Flight Howto!"),
	)
}

func TestHowtoIndexRendersEmptyStateWhenNoGuidesAvailable(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForHowtoTest(t, howtoToolFixtureForWeb(t, 0o755, `#!/bin/sh
cat <<'JSON'
{
  "success": true,
  "guides": []
}
JSON
`))

	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/howto", nil, http.StatusOK, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	testutil.AssertSelection(t, body, `[data-testid="howto-sidebar-empty"]`,
		testutil.HasText("No guides are currently available."),
	)
	testutil.AssertSelection(t, body, `[data-testid="howto-empty-state"]`,
		testutil.HasText("There are no user guides currently available. Return later or contact your administrator for more information."),
	)
}

func TestHowtoShowRendersSidebarSelectedGuideAndMarkdownHTML(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForHowtoTest(t, howtoToolFixtureForWeb(t, 0o755, "#!/bin/sh\n"+
		"if [ \"$1\" = \"list\" ]; then\n"+
		"cat <<'JSON'\n"+
		"{\n"+
		"  \"success\": true,\n"+
		"  \"guides\": [\n"+
		"    {\"index\": 1, \"title\": \"Base Guide A\"},\n"+
		"    {\"index\": 2, \"title\": \"Cluster Basics\"}\n"+
		"  ]\n"+
		"}\n"+
		"JSON\n"+
		"  exit 0\n"+
		"fi\n"+
		"cat <<'JSON'\n"+
		"{\n"+
		"  \"success\": true,\n"+
		"  \"guide\": {\n"+
		"    \"title\": \"Cluster Basics\",\n"+
		"    \"raw_markdown\": \"# Heading\\n\\nParagraph with [link](https://example.invalid).\\n\\n- Item one\\n- Item two\\n\\n> Quote\\n\\n`inline`\\n\\n```bash\\necho hi\\n```\\n\"\n"+
		"  }\n"+
		"}\n"+
		"JSON\n"))

	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/howto/2", nil, http.StatusOK, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	testutil.AssertSelection(t, body, `[data-testid="howto-guide-link--2"]`,
		testutil.HasText("Cluster Basics"),
		testutil.HasAttr("aria-current", "page"),
	)
	testutil.AssertSelection(t, body, `[data-testid="howto-guide-content"] h1`,
		testutil.HasText("Heading"),
	)
	testutil.AssertSelection(t, body, `[data-testid="howto-guide-content"] p a`,
		testutil.HasAttr("href", "https://example.invalid"),
		testutil.HasText("link"),
	)
	testutil.AssertSelection(t, body, `[data-testid="howto-guide-content"] ul li:first-child`,
		testutil.HasText("Item one"),
	)
	testutil.AssertSelection(t, body, `[data-testid="howto-guide-content"] blockquote p`,
		testutil.HasText("Quote"),
	)
	testutil.AssertSelection(t, body, `[data-testid="howto-guide-content"] pre code`,
		testutil.HasText("echo hi"),
	)
}

func TestHowtoShowRedirectsToIndexForInvalidGuideIndex(t *testing.T) {
	currentUser := currentUserForTest(t)
	setFlightRootForHowtoTest(t, howtoToolFixtureForWeb(t, 0o755, `#!/bin/sh
cat <<'JSON'
{
  "success": true,
  "guides": [
    {"index": 1, "title": "Base Guide A"}
  ]
}
JSON
`))

	resp, _ := testutil.RenderPage(t, newApp(), http.MethodGet, "/howto/2", nil, http.StatusSeeOther, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	if got := resp.Header.Get("location"); got != "/howto" {
		t.Fatalf("expected redirect to /howto, got %q", got)
	}
}

func TestHowtoIndexInvokesListJSONOnly(t *testing.T) {
	currentUser := currentUserForTest(t)
	argsFile := filepath.Join(t.TempDir(), "howto-args.txt")
	setFlightRootForHowtoTest(t, howtoToolFixtureForWeb(t, 0o755, "#!/bin/sh\n"+
		"printf '%s\\n' \"$@\" > \""+argsFile+"\"\n"+
		"cat <<'JSON'\n"+
		"{\n"+
		"  \"success\": true,\n"+
		"  \"guides\": []\n"+
		"}\n"+
		"JSON\n"))

	testutil.RenderPage(t, newApp(), http.MethodGet, "/howto", nil, http.StatusOK, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read command args fixture: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "list\n--format\njson" {
		t.Fatalf("expected list command args, got %q", got)
	}
}

func TestHowtoShowInvokesListBeforeShowJSON(t *testing.T) {
	currentUser := currentUserForTest(t)
	argsFile := filepath.Join(t.TempDir(), "howto-args.txt")
	setFlightRootForHowtoTest(t, howtoToolFixtureForWeb(t, 0o755, "#!/bin/sh\n"+
		"{\n"+
		"  printf '%s\\n' \"$@\"\n"+
		"  printf '%s\\n' '---'\n"+
		"} >> \""+argsFile+"\"\n"+
		"if [ \"$1\" = \"list\" ]; then\n"+
		"cat <<'JSON'\n"+
		"{\n"+
		"  \"success\": true,\n"+
		"  \"guides\": [\n"+
		"    {\"index\": 3, \"title\": \"Base Guide A\"}\n"+
		"  ]\n"+
		"}\n"+
		"JSON\n"+
		"  exit 0\n"+
		"fi\n"+
		"cat <<'JSON'\n"+
		"{\n"+
		"  \"success\": true,\n"+
		"  \"guide\": {\n"+
		"    \"title\": \"Base Guide A\",\n"+
		"    \"raw_markdown\": \"Body\\n\"\n"+
		"  }\n"+
		"}\n"+
		"JSON\n"))

	testutil.RenderPage(t, newApp(), http.MethodGet, "/howto/3", nil, http.StatusOK, testutil.WithSessionCookie(currentUser.Username, config.Session.Secret))

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("failed to read command args fixture: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "list\n--format\njson\n---\nshow\n--format\njson\n3\n---" {
		t.Fatalf("expected list then show command args, got %q", got)
	}
}

func setFlightRootForHowtoTest(t *testing.T, root string) {
	t.Helper()

	original := env.FlightRoot
	env.FlightRoot = root
	t.Cleanup(func() {
		env.FlightRoot = original
	})
}

func howtoToolFixtureForWeb(t *testing.T, mode os.FileMode, script string) string {
	t.Helper()

	root := t.TempDir()
	toolPath := filepath.Join(root, "usr", "lib", "flight-core", "flight-howto")
	if err := os.MkdirAll(filepath.Dir(toolPath), 0o755); err != nil {
		t.Fatalf("failed to create howto tool dir: %v", err)
	}
	if err := os.WriteFile(toolPath, []byte(script), mode); err != nil {
		t.Fatalf("failed to write howto tool fixture: %v", err)
	}
	synopsisDir := filepath.Join(root, "usr", "share", "doc", "tools")
	if err := os.MkdirAll(synopsisDir, 0o755); err != nil {
		t.Fatalf("failed to create synopsis dir: %v", err)
	}
	synopsisPath := filepath.Join(synopsisDir, "flight-howto")
	if err := os.WriteFile(synopsisPath, []byte("Learn about the Flight User Suite and using your cluster"), 0o644); err != nil {
		t.Fatalf("failed to write synopsis fixture: %v", err)
	}
	return root
}
