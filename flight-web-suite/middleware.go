package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func MethodOverrideMiddleware() echo.MiddlewareFunc {
	config := middleware.MethodOverrideConfig{
		Getter: middleware.MethodFromForm("_method"),
	}
	return middleware.MethodOverrideWithConfig(config)
}

func NewSessionMiddleware() echo.MiddlewareFunc {
	// TODO: Use a secret secret.
	sessionStore := sessions.NewCookieStore([]byte("totally-not-a-secret"))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("_session_store", sessionStore)
			return next(c)
		}
	}
}

func GetSession(c *echo.Context, name string) (*sessions.Session, error) {
	store, err := echo.ContextGet[sessions.Store](c, "_session_store")
	if err != nil {
		return nil, fmt.Errorf("failed to get session store: %w", err)
	}
	return store.Get(c.Request(), name)
}

func DeleteSession(c *echo.Context, name string) error {
	sess, err := GetSession(c, "session")
	if err != nil {
		return err
	}
	sess.Options.MaxAge = -1
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}
	return nil
}
