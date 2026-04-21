package main

import (
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/concertim/flight-user-suite/flight-web-suite/internal/testutil"
)

func TestSignInWorkflowGoodCredentials(t *testing.T) {
	setAuthenticatorPathForTest(t, filepath.Join("testdata", "authenticator_success.sh"))
	srv, client := testutil.SetupIntegrationServer(t, newApp())
	client.PostForm(srv.URL+"/sessions", url.Values{
		"username": {"bob"},
		"password": {"fakepassword"},
	})

	client.AssertRedirect(t, http.StatusSeeOther, "/")
	_, body := client.FollowRedirect()
	assertAuthenticated(t, body, "bob")
	testutil.AssertSelection(t, body, `div.flash.notice`,
		testutil.HasText("Successfully signed in"),
	)
}

func TestSignInWorkflowBadCredentials(t *testing.T) {
	setAuthenticatorPathForTest(t, filepath.Join("testdata", "authenticator_fail.sh"))
	srv, client := testutil.SetupIntegrationServer(t, newApp())
	_, body := client.PostForm(srv.URL+"/sessions", url.Values{
		"username": {"kate"},
		"password": {"fakepassword"},
	})

	client.AssertResponseCode(t, http.StatusOK)
	assertNotAuthenticated(t, body)
	testutil.AssertSelection(t, body, `div.flash.alert`,
		testutil.HasText("Invalid username or password"),
	)
	testutil.AssertSelection(t, body, `form[data-testid="new-session-form"] input[name="username"]`,
		testutil.HasAttr("value", "kate"),
	)
	testutil.AssertSelection(t, body, `form[data-testid="new-session-form"] input[name="password"]`,
		testutil.HasNoAttr("value"),
	)
}
