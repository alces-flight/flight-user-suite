package main

import (
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
	sessionStore := sessions.NewCookieStore([]byte(config.Session.Secret))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		Secure:   false,
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
