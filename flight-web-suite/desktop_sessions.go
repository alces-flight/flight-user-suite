package main

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/concertim/flight-user-suite/flight-web-suite/desktop"
	"github.com/labstack/echo/v5"
)

type desktopSessionCard struct {
	Name                 string
	DesktopType          string
	State                string
	Host                 string
	StartTimeText        string
	ActionLabel          string
	ActionTitle          string
	ActionEnabled        bool
	ActionPath           string
	ActionMethodOverride string
}

func indexDesktopSessionsHandler(c *echo.Context) error {
	if !IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/sessions")
	}
	if err := requireDesktopToolEnabled(); err != nil {
		return err
	}

	sessions, err := desktop.ListCommand(c.Request().Context(), env, CurrentUserName(c))
	if err != nil {
		return err
	}

	slices.SortFunc(sessions, func(a, b *desktop.Session) int {
		if byName := strings.Compare(a.Name, b.Name); byName != 0 {
			return byName
		}
		return a.CreatedAt.Compare(b.CreatedAt)
	})

	data := map[string]any{
		"Summary":         desktopSessionsSummary(sessions),
		"DesktopSessions": buildDesktopSessionCards(sessions),
	}
	return c.Render(http.StatusOK, "desktop/index", data)
}

func showDesktopSessionHandler(c *echo.Context) error {
	if !IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/sessions")
	}
	if err := requireDesktopToolEnabled(); err != nil {
		return err
	}

	response, err := desktop.ShowCommand(c.Request().Context(), env, CurrentUserName(c), c.Param("sessionName"))
	if err != nil {
		return err
	}
	sess, err := GetSession(c)
	if err != nil {
		return err
	}
	if !response.Success {
		if response.Reason == "not_found" {
			sess.AddFlash("Desktop session not found", "alert")
		} else {
			sess.AddFlash(fmt.Sprintf("Unexpected error finding desktop session: %s", response.Error), "alert")
		}
		SaveSession(c, sess)
		return c.Redirect(http.StatusSeeOther, "/desktop")
	}

	data := map[string]any{
		"layout":         "wide",
		"DesktopSession": response.Session,
	}
	return c.Render(http.StatusOK, "desktop/show", data)
}

func destroyDesktopSessionHandler(c *echo.Context) error {
	if !IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/sessions")
	}
	if err := requireDesktopToolEnabled(); err != nil {
		return err
	}

	response, err := desktop.KillCommand(c.Request().Context(), env, CurrentUserName(c), c.Param("sessionName"))
	if err != nil {
		return err
	}

	sess, err := GetSession(c)
	if err != nil {
		return err
	}
	if response.Success {
		sess.AddFlash(fmt.Sprintf("Desktop session '%s' terminated.", response.SessionName), "notice")
	} else {
		sess.AddFlash(fmt.Sprintf("Failed to terminate desktop session '%s': %s", response.SessionName, response.Error), "alert")
	}
	SaveSession(c, sess)
	return c.Redirect(http.StatusSeeOther, "/desktop")
}

func cleanDesktopSessionHandler(c *echo.Context) error {
	if !IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/sessions")
	}
	if err := requireDesktopToolEnabled(); err != nil {
		return err
	}

	response, err := desktop.CleanCommand(c.Request().Context(), env, CurrentUserName(c), c.Param("sessionName"))
	if err != nil {
		return err
	}

	sess, err := GetSession(c)
	if err != nil {
		return err
	}
	if response.Success {
		sess.AddFlash(fmt.Sprintf("Desktop session '%s' removed.", response.SessionName), "notice")
	} else {
		sess.AddFlash(fmt.Sprintf("Failed to remove desktop session '%s': %s", response.SessionName, response.Error), "alert")
	}
	SaveSession(c, sess)
	return c.Redirect(http.StatusSeeOther, "/desktop")
}

func buildDesktopSessionCards(sessions []*desktop.Session) []desktopSessionCard {
	cards := make([]desktopSessionCard, 0, len(sessions))
	for _, session := range sessions {
		actionLabel := ""
		actionTitle := ""
		actionEnabled := false
		actionPath := ""
		actionMethodOverride := ""
		switch session.State {
		case "active":
			actionLabel = "Terminate"
			actionTitle = ""
			actionEnabled = true
			actionPath = fmt.Sprintf("/desktop/%s", session.Name)
			actionMethodOverride = "DELETE"
		case "broken", "exited":
			actionLabel = "Remove"
			actionTitle = ""
			actionEnabled = true
			actionPath = fmt.Sprintf("/desktop/%s/clean", session.Name)
		case "remote":
			actionLabel = "Terminate"
			actionTitle = "Termination of remote sessions is not yet implemented."
		}

		cards = append(cards, desktopSessionCard{
			Name:                 session.Name,
			DesktopType:          session.DesktopType,
			State:                session.State,
			Host:                 session.Host,
			StartTimeText:        session.CreatedAt.Format("Mon 2 Jan 2006 15:04"),
			ActionLabel:          actionLabel,
			ActionTitle:          actionTitle,
			ActionEnabled:        actionEnabled,
			ActionPath:           actionPath,
			ActionMethodOverride: actionMethodOverride,
		})
	}
	return cards
}

func desktopSessionsSummary(sessions []*desktop.Session) string {
	count := len(sessions)
	runningCount := 0
	if count == 0 {
		return "You don't have any sessions currently running."
	}

	for _, s := range sessions {
		if s.State == "active" || s.State == "remote" {
			runningCount += 1
		}
	}
	if runningCount == count {
		if count == 1 {
			return "You have 1 desktop session currently running."
		}
		return fmt.Sprintf("You have %d desktop sessions currently running.", count)
	}

	nonRunningCount := count - runningCount
	nonRunningString := fmt.Sprintf("and %d stale sessions.", nonRunningCount)
	if nonRunningCount == 1 {
		nonRunningString = "and 1 stale session."
	}

	if runningCount == 1 {
		return fmt.Sprintf("You have 1 desktop session currently running %s", nonRunningString)
	}
	return fmt.Sprintf("You have %d desktop sessions currently running %s", runningCount, nonRunningString)
}
