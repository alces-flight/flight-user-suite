package main

import (
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/concertim/flight-user-suite/flight-web-suite/internal/testutil"
)

func TestHowtoShowInvalidIndexAddsFlashAndRedirects(t *testing.T) {
	currentUser := currentUserForTest(t)
	setAuthenticatorPathForTest(t, filepath.Join("testdata", "authenticator_success.sh"))
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

	srv, client := testutil.SetupIntegrationServer(t, newApp())
	client.PostForm(srv.URL+"/sessions", url.Values{
		"username": {currentUser.Username},
		"password": {"fakepassword"},
	})
	client.FollowRedirect()
	client.Get(srv.URL + "/howto/2")

	client.AssertRedirect(t, http.StatusSeeOther, "/howto")
	_, body := client.FollowRedirect()
	testutil.AssertSelection(t, body, `div.flash.alert`,
		testutil.HasText("Howto guide not found."),
	)
}

func TestHowtoShowNotFoundAddsFlashAndRedirects(t *testing.T) {
	currentUser := currentUserForTest(t)
	setAuthenticatorPathForTest(t, filepath.Join("testdata", "authenticator_success.sh"))
	setFlightRootForHowtoTest(t, howtoToolFixtureForWeb(t, 0o755, `#!/bin/sh
if [ "$1" = "list" ]; then
cat <<'JSON'
{
  "success": true,
  "guides": [
    {"index": 1, "title": "Base Guide A"}
  ]
}
JSON
  exit 0
fi
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

	srv, client := testutil.SetupIntegrationServer(t, newApp())
	client.PostForm(srv.URL+"/sessions", url.Values{
		"username": {currentUser.Username},
		"password": {"fakepassword"},
	})
	client.FollowRedirect()
	client.Get(srv.URL + "/howto/1")

	client.AssertRedirect(t, http.StatusSeeOther, "/howto")
	_, body := client.FollowRedirect()
	testutil.AssertSelection(t, body, `div.flash.alert`,
		testutil.HasText("Howto guide not found."),
	)
}
