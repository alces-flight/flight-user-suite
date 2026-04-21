package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/concertim/flight-user-suite/flight-web-suite/internal/testutil"
)

// Setup/teardown logic.
func TestMain(m *testing.M) {
	stateRoot, err := os.MkdirTemp("", "flight-web-suite-state-")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	flightStateRoot = stateRoot
	if err := os.Setenv("FLIGHT_STATE_ROOT", stateRoot); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	config, err = loadConfig()
	if err != nil {
		fmt.Println(err.Error())
		_ = os.RemoveAll(stateRoot)
		os.Exit(1)
	}
	config.Authenticator.Timeout = 1 * time.Second
	exitCode := m.Run()
	if !isRepoLocalStateRoot(stateRoot) {
		_ = os.RemoveAll(stateRoot)
	}
	os.Exit(exitCode)
}

func isRepoLocalStateRoot(path string) bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	return path == filepath.Join(cwd, "tmp")
}

func TestHomepageAnonymous(t *testing.T) {
	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/", nil, http.StatusOK)

	assertNotAuthenticated(t, body)
	assertToolCardAbsentHTML(t, body, "Flight Desktop")
	assertToolCardAbsentHTML(t, body, "Flight Howto")
}

func TestHomepageAuthenticated(t *testing.T) {
	setFlightRootForTest(t, "./testdata/flight_root") // Fake FLIGHT_ROOT with dummy tools enabled
	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/", nil, http.StatusOK, testutil.WithSessionCookie("ben", config.Session.Secret))

	assertAuthenticated(t, body, "ben")
	assertToolCardPresentHTML(t, body, "Flight Desktop", "/assets/images/desktop.png", "Access interactive desktop sessions")
	assertToolCardPresentHTML(t, body, "Flight Howto", "/assets/images/howto.png", "Learn about the Flight User Suite and using your cluster")
}

func assertToolCardPresentHTML(t *testing.T, body, title, imagePath, description string) {
	t.Helper()

	cardSelector := `div[data-testid="tool-card--` + title + `"]`
	testutil.AssertSelection(t, body, cardSelector+` h2`, testutil.HasText(title))
	testutil.AssertSelection(t, body, cardSelector+` p`, testutil.HasText(description))
	testutil.AssertSelection(t, body, cardSelector+` img`,
		testutil.HasAttr("src", imagePath),
		testutil.HasAttr("alt", title+" logo"),
	)
}

func assertToolCardAbsentHTML(t *testing.T, body, title string) {
	t.Helper()

	cardSelector := `div[data-testid="tool-card--` + title + `"]`
	testutil.AssertNoSelection(t, body, cardSelector)
}

func assertNotAuthenticated(t *testing.T, body string) {
	t.Helper()

	testutil.AssertSelection(t, body, `a[data-testid="sign-in-link"]`,
		testutil.HasAttr("href", "/sessions"),
		testutil.HasText("Login"),
	)
	testutil.AssertNoSelection(t, body, `[data-testid="signed-in-message"]`)
	testutil.AssertNoSelection(t, body, `[data-testid="logout-form"]`)
}

func assertAuthenticated(t *testing.T, body, username string) {
	t.Helper()

	testutil.AssertSelection(t, body, `[data-testid="signed-in-message"]`,
		testutil.HasText("You are signed in as "+username),
	)
	testutil.AssertSelection(t, body, `form[data-testid="logout-form"]`,
		testutil.HasAttr("action", "sessions"),
		testutil.HasAttr("method", "post"),
	)
	testutil.AssertSelection(t, body, `form[data-testid="logout-form"] input[name="_method"]`,
		testutil.HasAttr("value", "DELETE"),
	)
	testutil.AssertSelection(t, body, `form[data-testid="logout-form"] button`,
		testutil.HasText("Logout"),
	)
	testutil.AssertNoSelection(t, body, `[data-testid="sign-in-link"]`)
}

func setFlightRootForTest(t *testing.T, path string) {
	t.Helper()

	origPath := flightRoot
	flightRoot = path
	t.Cleanup(func() {
		flightRoot = origPath
	})
}
