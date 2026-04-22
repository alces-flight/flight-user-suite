package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v5"
)

const sessionContextKey = "_session"

func Flashes(c *echo.Context, flavour string) []string {
	sess, err := GetSession(c)
	if err != nil {
		return nil
	}
	var flashes []string
	for _, ef := range sess.Flashes(flavour) {
		switch ef := ef.(type) {
		case string:
			flashes = append(flashes, ef)
		case fmt.Stringer:
			flashes = append(flashes, ef.String())
		}
	}
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return flashes
	}
	return flashes
}

func GetSession(c *echo.Context) (*sessions.Session, error) {
	if sess, ok := c.Get(sessionContextKey).(*sessions.Session); ok {
		return sess, nil
	}

	name := "session"
	store, err := echo.ContextGet[sessions.Store](c, "_session_store")
	if err != nil {
		return nil, fmt.Errorf("failed to get session store: %w", err)
	}
	sess, err := store.Get(c.Request(), name)
	if err != nil {
		if multiErr, ok := errors.AsType[securecookie.MultiError](err); ok && multiErr.IsDecode() {
			// If we can't decode the session, use an empty session for the
			// rest of this request.  We cache the session to ensure subsequent
			// calls to `GetSession()` return it without logging again.
			CacheSession(c, sess)
			c.Logger().Warn("error decoding session: using empty session", "err", multiErr)
			return sess, nil
		} else {
			return sess, err
		}
	}
	return sess, err
}

func DeleteSession(c *echo.Context) error {
	sess, err := GetSession(c)
	if err != nil {
		return err
	}
	sess.Options.MaxAge = -1
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}
	return nil
}

func SaveSession(c *echo.Context, sess *sessions.Session) {
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		c.Logger().Warn("Error saving session", "err", err)
	}
}

func CacheSession(c *echo.Context, sess *sessions.Session) {
	c.Set(sessionContextKey, sess)
}

func IsLoggedIn(c *echo.Context) bool {
	sess, err := GetSession(c)
	if err != nil {
		return false
	}
	if username, ok := sess.Values["username"].(string); ok && username != "" {
		return true
	}
	return false
}

func CurrentUserName(c *echo.Context) string {
	sess, err := GetSession(c)
	if err != nil {
		return ""
	}
	if username, ok := sess.Values["username"].(string); ok && username != "" {
		return username
	}
	return ""
}

func AddCommonData(c *echo.Context, data map[string]any) map[string]any {
	if data == nil {
		data = make(map[string]any)
	}
	data["CurrentUserName"] = CurrentUserName(c)
	data["flashes"] = map[string]string{
		"notice": strings.Join(Flashes(c, "notice"), ", "),
		"alert":  strings.Join(Flashes(c, "alert"), ", "),
	}
	return data
}
