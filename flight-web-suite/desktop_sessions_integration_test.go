package main

import (
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/concertim/flight-user-suite/flight-web-suite/internal/testutil"
)

func TestDestroyDesktopSessionSuccessAddsFlashAndRedirects(t *testing.T) {
	currentUser := currentUserForTest(t)
	setAuthenticatorPathForTest(t, filepath.Join("testdata", "authenticator_success.sh"))
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
if [ "$1" = "list" ]; then
  echo '[]'
  exit 0
fi
cat <<'JSON'
{
  "success": true,
  "session_name": "alpha"
}
JSON
`))

	srv, client := testutil.SetupIntegrationServer(t, newApp())
	client.PostForm(srv.URL+"/sessions", url.Values{
		"username": {currentUser.Username},
		"password": {"fakepassword"},
	})
	client.FollowRedirect()
	client.PostForm(srv.URL+"/desktop/alpha", url.Values{
		"_method": {"DELETE"},
	})

	client.AssertRedirect(t, http.StatusSeeOther, "/desktop")
	_, body := client.FollowRedirect()
	testutil.AssertSelection(t, body, `div.flash.notice`,
		testutil.HasText("Desktop session 'alpha' terminated."),
	)
}

func TestDestroyDesktopSessionFailureAddsFlashAndRedirects(t *testing.T) {
	currentUser := currentUserForTest(t)
	setAuthenticatorPathForTest(t, filepath.Join("testdata", "authenticator_success.sh"))
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
if [ "$1" = "list" ]; then
  echo '[]'
  exit 0
fi
cat <<'JSON'
{
  "success": false,
  "session_name": "alpha",
  "error": "Desktop session 'alpha' is not active.",
  "reason": "not_active"
}
JSON
exit 1
`))

	srv, client := testutil.SetupIntegrationServer(t, newApp())
	client.PostForm(srv.URL+"/sessions", url.Values{
		"username": {currentUser.Username},
		"password": {"fakepassword"},
	})
	client.FollowRedirect()
	client.PostForm(srv.URL+"/desktop/alpha", url.Values{
		"_method": {"DELETE"},
	})

	client.AssertRedirect(t, http.StatusSeeOther, "/desktop")
	_, body := client.FollowRedirect()
	testutil.AssertSelection(t, body, `div.flash.alert`,
		testutil.HasText("Failed to terminate desktop session 'alpha': Desktop session 'alpha' is not active."),
	)
}

func TestCleanDesktopSessionSuccessAddsFlashAndRedirects(t *testing.T) {
	currentUser := currentUserForTest(t)
	setAuthenticatorPathForTest(t, filepath.Join("testdata", "authenticator_success.sh"))
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
if [ "$1" = "list" ]; then
  echo '[]'
  exit 0
fi
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

	srv, client := testutil.SetupIntegrationServer(t, newApp())
	client.PostForm(srv.URL+"/sessions", url.Values{
		"username": {currentUser.Username},
		"password": {"fakepassword"},
	})
	client.FollowRedirect()
	client.PostForm(srv.URL+"/desktop/alpha/clean", url.Values{})

	client.AssertRedirect(t, http.StatusSeeOther, "/desktop")
	_, body := client.FollowRedirect()
	testutil.AssertSelection(t, body, `div.flash.notice`,
		testutil.HasText("Desktop session 'alpha' removed."),
	)
}

func TestCleanDesktopSessionFailureAddsFlashAndRedirects(t *testing.T) {
	currentUser := currentUserForTest(t)
	setAuthenticatorPathForTest(t, filepath.Join("testdata", "authenticator_success.sh"))
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
if [ "$1" = "list" ]; then
  echo '[]'
  exit 0
fi
cat <<'JSON'
{
  "success": false,
  "results": [
    {
      "success": false,
      "session_name": "alpha",
      "error": "Desktop session 'alpha' is active.",
      "reason": "active"
    }
  ]
}
JSON
exit 1
`))

	srv, client := testutil.SetupIntegrationServer(t, newApp())
	client.PostForm(srv.URL+"/sessions", url.Values{
		"username": {currentUser.Username},
		"password": {"fakepassword"},
	})
	client.FollowRedirect()
	client.PostForm(srv.URL+"/desktop/alpha/clean", url.Values{})

	client.AssertRedirect(t, http.StatusSeeOther, "/desktop")
	_, body := client.FollowRedirect()
	testutil.AssertSelection(t, body, `div.flash.alert`,
		testutil.HasText("Failed to remove desktop session 'alpha': Desktop session 'alpha' is active."),
	)
}
