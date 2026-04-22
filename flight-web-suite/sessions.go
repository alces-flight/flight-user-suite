package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"

	"github.com/labstack/echo/v5"
)

func newSessionHandler(c *echo.Context) error {
	if IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/")
	}
	return c.Render(http.StatusOK, "sessions/new", AddCommonData(c, nil))
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
		return c.Render(http.StatusOK, "sessions/new", AddCommonData(c, nil))
	}
	ok, err := authenticate(c.Request().Context(), username, password)
	if err != nil {
		return err
	}

	if ok {
		sess.Values["username"] = username
		sess.AddFlash("Successfully signed in", "notice")
		SaveSession(c, sess)
		return c.Redirect(http.StatusSeeOther, "/")
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
	return c.Redirect(http.StatusSeeOther, "/")
}

func authenticate(ctx context.Context, username, password string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, config.Authenticator.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, authenticatorPath, username)
	cmd.Env = append(os.Environ(), "FLIGHT_WEB_SUITE_PAM_SERVICE="+config.Authenticator.PAMService)
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
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			return false, err
		}
	}
	return cmd.ProcessState.Success(), nil
}
