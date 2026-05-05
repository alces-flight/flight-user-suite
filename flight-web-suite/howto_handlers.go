package main

import (
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strconv"

	"github.com/concertim/flight-user-suite/flight-web-suite/howto"
	"github.com/concertim/flight-user-suite/flight/toolset"
	"github.com/labstack/echo/v5"
)

type howtoPageGuide struct {
	Index      int
	Title      string
	URL        string
	IsSelected bool
}

func indexHowtoHandler(c *echo.Context) error {
	if !IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/sessions")
	}
	if err := requireHowtoToolEnabled(); err != nil {
		return err
	}

	response, err := howto.ListCommand(c.Request().Context(), env, CurrentUserName(c))
	if err != nil {
		return err
	}
	if !response.Success {
		return fmt.Errorf("listing howtos: %s", response.Error)
	}

	return renderHowtoPage(c, response.Guides, 0, "")
}

func showHowtoHandler(c *echo.Context) error {
	if !IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/sessions")
	}
	if err := requireHowtoToolEnabled(); err != nil {
		return err
	}

	response, err := howto.ListCommand(c.Request().Context(), env, CurrentUserName(c))
	if err != nil {
		return err
	}
	if !response.Success {
		return fmt.Errorf("listing howtos: %s", response.Error)
	}
	if len(response.Guides) == 0 {
		return renderHowtoPage(c, response.Guides, 0, "")
	}

	index, err := strconv.Atoi(c.Param("index"))
	if err != nil || index <= 0 || !containsGuideIndex(response.Guides, index) {
		return redirectHowtoIndexAlert(c, "Howto guide not found.")
	}

	showResponse, err := howto.ShowCommand(c.Request().Context(), env, CurrentUserName(c), index)
	if err != nil {
		return err
	}
	if !showResponse.Success {
		if showResponse.Reason == "not_found" || showResponse.Reason == "invalid_index" {
			return redirectHowtoIndexAlert(c, "Howto guide not found.")
		}
		return fmt.Errorf("showing howto %d: %s", index, showResponse.Error)
	}

	rendered, err := renderMarkdown(showResponse.Guide.RawMarkdown)
	if err != nil {
		return err
	}
	return renderHowtoPage(c, response.Guides, index, rendered)
}

func requireHowtoToolEnabled() error {
	tool, err := toolset.GetTool(env.FlightRoot, "howto")
	if err != nil || !tool.Enabled {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Flight Howto is not enabled")
	}
	return nil
}

func redirectHowtoIndexAlert(c *echo.Context, message string) error {
	sess, err := GetSession(c)
	if err != nil {
		return err
	}
	sess.AddFlash(message, "alert")
	SaveSession(c, sess)
	return c.Redirect(http.StatusSeeOther, "/howto")
}

func containsGuideIndex(guides []howto.GuideSummary, index int) bool {
	return slices.ContainsFunc(guides, func(guide howto.GuideSummary) bool {
		return guide.Index == index
	})
}

func renderHowtoPage(c *echo.Context, guides []howto.GuideSummary, selectedIndex int, content template.HTML) error {
	pageGuides := make([]howtoPageGuide, 0, len(guides))
	for _, guide := range guides {
		pageGuides = append(pageGuides, howtoPageGuide{
			Index:      guide.Index,
			Title:      guide.Title,
			URL:        fmt.Sprintf("/howto/%d", guide.Index),
			IsSelected: guide.Index == selectedIndex,
		})
	}

	data := map[string]any{
		"layout":          "howto",
		"HasGuides":       len(pageGuides) > 0,
		"Guides":          pageGuides,
		"GuideContent":    content,
		"HasSelection":    selectedIndex > 0,
		"SelectedIndex":   selectedIndex,
		"ShowEmptyState":  len(pageGuides) == 0,
		"ShowSelectState": len(pageGuides) > 0 && selectedIndex == 0,
	}
	return c.Render(http.StatusOK, "howto/index", data)
}
