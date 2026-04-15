package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/containerd/fifo"
	"github.com/labstack/echo/v5"
)

func newSessionHandler(c *echo.Context) error {
	sess, err := GetSession(c, "session")
	if err != nil {
		return err
	}
	username, ok := sess.Values["username"].(string)
	if !ok || username == "" {
		fmt.Printf("err: %v\n", err)
		return c.HTML(
			http.StatusOK,
			`
<html>
<head></head>
<body>
	<form action="sessions" method="post">
		<input name="username" id="username" type="text" placeholder="Enter cluster username" />
		<input name="password" id="password" type="password" placeholder="Enter cluster password" />
		<input type="submit" value="Login" />
	</form>
</body>
</html>
`,
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
	c.Logger().Info("Doing auth")
	// TODO:
	// * Extract username and password.
	// * Launch external program pass username and password via pipe.
	// * Either redirect with error or create session cookie.

	sess, err := GetSession(c, "session")
	if err != nil {
		c.Logger().Error("GetSession", "err", err.Error())
		return err
	}

	fifoPath, fi, err := openFifo(c)
	if err != nil {
		c.Logger().Error("openFifo", "err", err.Error())
		return err
	}
	defer fi.Close() // nolint:errcheck

	c.Logger().Info("Building command")
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(
		ctx,
		filepath.Join(flightRoot, "usr", "libexec", "web-suite", "authenticate.py"),
		[]string{
			c.FormValue("username"),
			fifoPath,
		}...,
	)
	fmt.Printf("cmd: %v\n", cmd)
	fmt.Printf("cmd.Args: %v\n", cmd.Args)

	c.Logger().Info("Running command")
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	if err := cmd.Start(); err != nil {
		return err
	}

	c.Logger().Info("Writing to FIFO")
	n, err := fi.Write([]byte(c.FormValue("password")))
	fmt.Printf("n: %v\n", n)
	if err != nil {
		c.Logger().Error("fi.Write", "err", err.Error())
		return err
	}
	c.Logger().Info("Closing FIFO")
	err = fi.Close()
	if err != nil {
		c.Logger().Error("fi.Close", "err", err.Error())
		return err
	}

	c.Logger().Info("Waiting on cmd")
	err = cmd.Wait()
	if err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			c.Logger().Error("cmd.Wait", "err", err.Error())
			return err
		}
	}
	if cmd.ProcessState.Success() {
		sess.Values["username"] = c.FormValue("username")
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			c.Logger().Error("sess.Save", "err", err.Error())
			return err
		}
		sess.AddFlash("Successfully signed in", "info")
	} else {
		sess.AddFlash("Invalid username or password", "error")
	}

	return c.Redirect(http.StatusFound, "/")
}

func openFifo(c *echo.Context) (string, io.ReadWriteCloser, error) {
	c.Logger().Info("Creating tmpdir")
	tmpDir, err := os.MkdirTemp("", "flight-web-")
	if err != nil {
		c.Logger().Error("MkdirTemp", "err", err.Error())
		return "", nil, err
	}
	c.Logger().Info("Creating fifo")
	fifoPath := filepath.Join(tmpDir, "passwd")
	fi, err := fifo.OpenFifo(c.Request().Context(), fifoPath, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_NONBLOCK, 0o700)
	if err != nil {
		c.Logger().Error("fifo.OpenFifo", "err", err.Error())
		return "", nil, err
	}
	return fifoPath, fi, nil
}

func destroySessionHandler(c *echo.Context) error {
	err := DeleteSession(c, "session")
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/")
}
