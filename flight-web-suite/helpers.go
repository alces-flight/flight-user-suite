package main

import (
	"fmt"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v5"
)

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
	name := "session"
	store, err := echo.ContextGet[sessions.Store](c, "_session_store")
	if err != nil {
		return nil, fmt.Errorf("failed to get session store: %w", err)
	}
	return store.Get(c.Request(), name)
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

func LoggedIn(c *echo.Context) bool {
	sess, err := GetSession(c)
	if err != nil {
		return false
	}
	if username, ok := sess.Values["username"].(string); ok && username != "" {
		return true
	}
	return false
}

func CurrentUser(c *echo.Context) string {
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
	data["CurrentUser"] = CurrentUser(c)
	data["flashes"] = map[string]string{
		"notice": strings.Join(Flashes(c, "notice"), ", "),
		"alert":  strings.Join(Flashes(c, "alert"), ", "),
	}
	return data
}
