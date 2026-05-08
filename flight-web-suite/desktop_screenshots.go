package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v5"
)

// showScreenshotHandler serves the screenshot for the named session.  If there
// is any error or inability to do so, the placeholder image is served instead.
func showScreenshotHandler(c *echo.Context) error {
	placeholderImagePath := filepath.Join(getDirectory("assets"), "images/placeholder.jpg")
	if !IsLoggedIn(c) {
		return c.File(placeholderImagePath)
	}
	if err := requireDesktopToolEnabled(); err != nil {
		return c.File(placeholderImagePath)
	}
	cli := desktopCli(c.Logger())
	response, err := cli.ShowCommand(c.Request().Context(), CurrentUserName(c), c.Param("sessionName"))
	if err != nil {
		return c.File(placeholderImagePath)
	}
	if !response.Success || response.Session.ScreenshotPath == "" {
		return c.File(placeholderImagePath)
	}
	fi, err := os.Stat(response.Session.ScreenshotPath)
	if err != nil {
		return c.File(placeholderImagePath)
	}

	// [echo.Context.File] won't work here as we can't guarantee that the path
	// is inside the [echo.Echo.Filesystem]. In fact it likely isn't.
	f, err := os.Open(response.Session.ScreenshotPath)
	if err != nil {
		return c.File(placeholderImagePath)
	}
	defer f.Close()

	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
	return nil
}
