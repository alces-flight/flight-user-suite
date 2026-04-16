package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
)

func flashes(c *echo.Context, flavour string) []string {
	sess, err := GetSession(c, "session")
	if err != nil {
		return nil
	}
	var errorFlashes []string
	for _, ef := range sess.Flashes(flavour) {
		switch ef := ef.(type) {
		case string:
			errorFlashes = append(errorFlashes, ef)
		case fmt.Stringer:
			errorFlashes = append(errorFlashes, ef.String())
		}
	}
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return errorFlashes
	}
	return errorFlashes
}

func newSessionHandler(c *echo.Context) error {
	sess, err := GetSession(c, "session")
	if err != nil {
		return err
	}
	username, ok := sess.Values["username"].(string)
	if !ok || username == "" {
		return c.HTML(
			http.StatusOK,
			fmt.Sprintf(`
<html>
<head></head>
<body>
    <div>%s</div>
	<form action="sessions" method="post">
		<input name="username" id="username" type="text" placeholder="Enter cluster username" />
		<input name="password" id="password" type="password" placeholder="Enter cluster password" />
		<input type="submit" value="Login" />
	</form>
</body>
</html>
`,
				strings.Join(flashes(c, "error"), ", "),
			),
		)
	}

	return c.HTML(
		http.StatusOK,
		fmt.Sprintf(`
		<html>
			<head></head>
			<body>
				<div>You are logged in as %s</div>
				<div>
					<form action="sessions" method="post">
						<input type="hidden" name="_method" value="DELETE" />
						<input type="submit" value="Logout" />
					</form>
				</div>
			</body>
		</html>
		`,
			sess.Values["username"],
		),
	)
}

func createSessionHandler(c *echo.Context) error {
	sess, err := GetSession(c, "session")
	if err != nil {
		return err
	}

	username := c.FormValue("username")
	password := c.FormValue("password")
	if username == "" || password == "" {
		sess.AddFlash("Username and/or password not provided", "error")
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return err
		}
		return c.Redirect(http.StatusFound, "/")
	}
	ok, err := authenticate(c.Request().Context(), username, password)
	if err != nil {
		return err
	}

	if ok {
		sess.Values["username"] = username
		sess.AddFlash("Successfully signed in", "info")
	} else {
		sess.AddFlash("Invalid username or password", "error")
	}
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

func destroySessionHandler(c *echo.Context) error {
	err := DeleteSession(c, "session")
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
