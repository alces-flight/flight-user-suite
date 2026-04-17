package main

import (
	"context"
	"errors"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v5"
)

func newSessionHandler(c *echo.Context) error {
	if LoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/")
	}
	return c.Render(http.StatusOK, "sessions/new", map[string]string{})
}

func createSessionHandler(c *echo.Context) error {
	sess, err := GetSession(c)
	if err != nil {
		return err
	}

	username := c.FormValue("username")
	password := c.FormValue("password")
	if username == "" || password == "" {
		sess.AddFlash("Username and/or password not provided", "alert")
		SaveSession(c, sess)
		return c.Redirect(http.StatusFound, "/")
	}
	ok, err := authenticate(c.Request().Context(), username, password)
	if err != nil {
		return err
	}

	if ok {
		sess.Values["username"] = username
		sess.AddFlash("Successfully signed in", "notice")
		SaveSession(c, sess)
		return c.Redirect(http.StatusFound, "/")
	} else {
		sess.AddFlash("Invalid username or password", "alert")
		SaveSession(c, sess)
		data := map[string]any{"username": username}
		return c.Render(http.StatusOK, "sessions/new", AddCommonData(c, data))
	}
}

func destroySessionHandler(c *echo.Context) error {
	err := DeleteSession(c)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/")
}

func authenticate(ctx context.Context, username, password string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	authenticator := filepath.Join(flightRoot, "usr", "libexec", "web-suite", "authenticate.py")
	cmd := exec.CommandContext(ctx, authenticator, username)
	pipe, err := cmd.StdinPipe()
	if err != nil {
		return false, err
	}
	if err := cmd.Start(); err != nil {
		return false, err
	}
	_, err = pipe.Write([]byte(password))
	if err != nil {
		return false, err
	}
	pipe.Close() // nolint:errcheck
	err = cmd.Wait()
	if err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			return false, err
		}
	}
	return cmd.ProcessState.Success(), nil
}
