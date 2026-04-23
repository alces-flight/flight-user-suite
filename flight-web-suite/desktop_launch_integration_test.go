package main

import (
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/concertim/flight-user-suite/flight-web-suite/internal/testutil"
)

func TestCreateDesktopSessionSuccessAddsFlashAndRedirects(t *testing.T) {
	currentUser := currentUserForTest(t)
	setAuthenticatorPathForTest(t, filepath.Join("testdata", "authenticator_success.sh"))
	setFlightRootForDesktopTest(t, desktopToolFixture(t, 0o755, `#!/bin/sh
if [ "$1" = "list" ]; then
  echo '[]'
  exit 0
fi
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

	srv, client := testutil.SetupIntegrationServer(t, newApp())
	client.PostForm(srv.URL+"/sessions", url.Values{
		"username": {currentUser.Username},
		"password": {"fakepassword"},
	})
	client.FollowRedirect()
	client.PostForm(srv.URL+"/desktop", url.Values{
		"desktop_type": {"xterm"},
		"name":         {"named-session"},
		"geometry":     {"1024x768"},
	})

	client.AssertRedirect(t, http.StatusSeeOther, "/desktop")
	_, body := client.FollowRedirect()
	testutil.AssertSelection(t, body, `div.flash.notice`,
		testutil.HasText("Desktop session 'named-session' launched."),
	)
}

func TestCreateDesktopSessionFailureAddsFlashAndRedisplaysForm(t *testing.T) {
	currentUser := currentUserForTest(t)
	setAuthenticatorPathForTest(t, filepath.Join("testdata", "authenticator_success.sh"))
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
  "success": false,
  "session_name": "named-session",
  "error": "Desktop startup failed.",
  "reason": "start_failed"
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
	_, body := client.PostForm(srv.URL+"/desktop", url.Values{
		"desktop_type": {"xterm"},
		"name":         {"named-session"},
		"geometry":     {"1024x768"},
	})

	client.AssertResponseCode(t, http.StatusUnprocessableEntity)
	testutil.AssertSelection(t, body, `div.flash.alert`,
		testutil.HasText("Desktop startup failed."),
	)
}
