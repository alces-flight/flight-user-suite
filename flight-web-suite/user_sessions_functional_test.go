package main

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/concertim/flight-user-suite/flight-web-suite/internal/testutil"
)

func TestNewSessionPageDisplaysFormForAnonymous(t *testing.T) {
	_, body := testutil.RenderPage(t, newApp(), http.MethodGet, "/sessions", nil, http.StatusOK)

	assertNotAuthenticated(t, body)
	testutil.AssertSelection(t, body, `form[data-testid="new-session-form"]`,
		testutil.HasAttr("action", "sessions"),
		testutil.HasAttr("method", "post"),
	)
	testutil.AssertSelection(t, body, `form[data-testid="new-session-form"] input[name="username"]`)
	testutil.AssertSelection(t, body, `form[data-testid="new-session-form"] input[name="password"]`)
}

func TestNewSessionPageRedirectsForAuthenticated(t *testing.T) {
	resp, _ := testutil.RenderPage(t, newApp(), http.MethodGet, "/sessions", nil, http.StatusSeeOther, testutil.WithSessionCookie("ben", config.Session.Secret))

	if resp.Header.Get("location") != "/" {
		t.Errorf("Expected new sessions page to redirect to '/' for authenticated users")
	}
}

func TestCreateNewSessionNoData(t *testing.T) {
	data := strings.NewReader("")
	_, body := testutil.RenderPage(t, newApp(), http.MethodPost, "/sessions", data, http.StatusOK, testutil.WithContentType("application/x-www-form-urlencoded"))

	assertNotAuthenticated(t, body)
	testutil.AssertSelection(t, body, `div.flash.alert`,
		testutil.HasText("Username and/or password not provided"),
	)
}

func TestCreateNewSessionBadData(t *testing.T) {
	setAuthenticatorPathForTest(t, filepath.Join("testdata", "authenticator_fail.sh"))
	data := strings.NewReader("username=bob&password=fakepassword")
	_, body := testutil.RenderPage(t, newApp(), http.MethodPost, "/sessions", data, http.StatusOK, testutil.WithContentType("application/x-www-form-urlencoded"))

	assertNotAuthenticated(t, body)
	testutil.AssertSelection(t, body, `div.flash.alert`,
		testutil.HasText("Invalid username or password"),
	)
	testutil.AssertSelection(t, body, `form[data-testid="new-session-form"] input[name="username"]`,
		testutil.HasAttr("value", "bob"),
	)
	testutil.AssertSelection(t, body, `form[data-testid="new-session-form"] input[name="password"]`,
		testutil.HasNoAttr("value"),
	)
}

func TestCreateNewSessionGoodData(t *testing.T) {
	setAuthenticatorPathForTest(t, filepath.Join("testdata", "authenticator_success.sh"))
	data := strings.NewReader("username=bob&password=fakepassword")
	resp, _ := testutil.RenderPage(t, newApp(), http.MethodPost, "/sessions", data, http.StatusSeeOther, testutil.WithContentType("application/x-www-form-urlencoded"))

	if resp.Header.Get("location") != "/" {
		t.Errorf("Expected new sessions page to redirect to '/' for authenticated users")
	}
}

func TestCreateNewSessionInternalErrors(t *testing.T) {
	testCases := []struct {
		name string
		path string
	}{
		{
			name: "timeout",
			path: filepath.Join("testdata", "authenticator_timeout.sh"),
		},
		{
			name: "missing",
			path: filepath.Join("testdata", "authenticator_missing.sh"),
		},
		{
			name: "not executable",
			path: filepath.Join("testdata", "authenticator_not_executable.sh"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setAuthenticatorPathForTest(t, tc.path)
			data := strings.NewReader("username=bob&password=fakepassword")
			resp, body := testutil.RenderPage(t, newApp(), http.MethodPost, "/sessions", data, http.StatusInternalServerError, testutil.WithContentType("application/x-www-form-urlencoded"))

			if location := resp.Header.Get("location"); location != "" {
				t.Errorf("expected no redirect when authenticator %s, got location %q", tc.name, location)
			}
			if !strings.Contains(body, "Internal Server Error") {
				t.Errorf("expected generic internal server error response body, got %q", body)
			}
		})
	}
}

func setAuthenticatorPathForTest(t *testing.T, path string) {
	t.Helper()

	origPath := authenticatorPath
	authenticatorPath = path
	t.Cleanup(func() {
		authenticatorPath = origPath
	})
}
