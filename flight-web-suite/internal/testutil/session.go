package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
)

// SessionCookie returns a signed session cookie containing the provided username.
func SessionCookie(t *testing.T, username string) *http.Cookie {
	t.Helper()

	store := sessions.NewCookieStore([]byte("totally-not-a-secret"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	sess, err := store.Get(req, "session")
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	sess.Values["username"] = username
	if err := sess.Save(req, rec); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	res := rec.Result()
	defer res.Body.Close()

	for _, cookie := range res.Cookies() {
		if cookie.Name == "session" {
			return cookie
		}
	}
	t.Fatal("session cookie not found")
	return nil
}

// AddSessionCookie adds a signed session cookie for username to req.
func AddSessionCookie(t *testing.T, req *http.Request, username string) {
	t.Helper()
	req.AddCookie(SessionCookie(t, username))
}

func WithSessionCookie(username string) RequestOption {
	return func(t *testing.T, req *http.Request) {
		t.Helper()
		AddSessionCookie(t, req, username)
	}
}
