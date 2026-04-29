package main

import (
	"fmt"
	"net/http"
	"net/url"

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

func NewWSProxyMiddleware() echo.MiddlewareFunc {
	return middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: &WebsockifyBalancer{},
		Skipper:  middleware.DefaultSkipper,
	})
}

// WebsockifyBalancer dynamically determines the upstream server by examining
// the incoming request.  It's more of a router than a balancer.
type WebsockifyBalancer struct{}

func (b *WebsockifyBalancer) Next(c *echo.Context) (*middleware.ProxyTarget, error) {
	host := c.QueryParam("host")
	port := c.QueryParam("port")
	targetURL, err := url.Parse(fmt.Sprintf("http://%s:%s/", host, port))
	c.Logger().Info("proxying websockify request", "target", targetURL.String())
	if err != nil {
		return nil, err
	}

	return &middleware.ProxyTarget{
		URL: targetURL,
	}, nil
}

// Implement remaining [middleware.ProxyBalancer] interface.
func (b *WebsockifyBalancer) AddTarget(target *middleware.ProxyTarget) bool { return true }
func (b *WebsockifyBalancer) RemoveTarget(name string) bool                 { return true }
