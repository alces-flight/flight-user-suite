package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/concertim/flight-user-suite/flight/toolset"
	"github.com/labstack/echo/v5"
)

type desktopSession struct {
	Name        string    `json:"name"`
	DesktopType string    `json:"desktop_type"`
	State       string    `json:"state"`
	Host        string    `json:"host"`
	CreatedAt   time.Time `json:"created_at"`
}

type desktopSessionCard struct {
	Name          string
	DesktopType   string
	State         string
	Host          string
	StartTimeText string
	ActionLabel   string
	ActionTitle   string
	ActionEnabled bool
}

func indexDesktopSessionsHandler(c *echo.Context) error {
	if !IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/sessions")
	}

	tool, err := toolset.GetTool(env.FlightRoot, "desktop")
	if err != nil || !tool.Enabled {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Flight Desktop is not enabled")
	}

	sessions, err := desktopListCommand(c.Request().Context(), CurrentUserName(c))
	if err != nil {
		return err
	}

	slices.SortFunc(sessions, func(a, b desktopSession) int {
		if byName := strings.Compare(a.Name, b.Name); byName != 0 {
			return byName
		}
		return a.CreatedAt.Compare(b.CreatedAt)
	})

	data := map[string]any{
		"Summary":         desktopSessionsSummary(sessions),
		"DesktopSessions": buildDesktopSessionCards(sessions),
	}
	return c.Render(http.StatusOK, "desktop/index", AddCommonData(c, data))
}

func destroyDesktopSessionHandler(c *echo.Context) error {
	if !IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/sessions")
	}

	tool, err := toolset.GetTool(env.FlightRoot, "desktop")
	if err != nil || !tool.Enabled {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Flight Desktop is not enabled")
	}

	response, err := desktopKillCommand(c.Request().Context(), CurrentUserName(c), c.Param("sessionName"))
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

func buildDesktopSessionCards(sessions []desktopSession) []desktopSessionCard {
	cards := make([]desktopSessionCard, 0, len(sessions))
	for _, session := range sessions {
		actionLabel := "Clean"
		actionTitle := "Cleaning desktop sessions is not yet implemented."
		actionEnabled := false
		switch session.State {
		case "active":
			actionLabel = "Terminate"
			actionTitle = ""
			actionEnabled = true
		case "remote":
			actionLabel = "Terminate"
			actionTitle = "Termination of remote sessions is not yet implemented."
		}

		cards = append(cards, desktopSessionCard{
			Name:          session.Name,
			DesktopType:   session.DesktopType,
			State:         session.State,
			Host:          session.Host,
			StartTimeText: session.CreatedAt.Format("Mon 2 Jan 2006 15:04"),
			ActionLabel:   actionLabel,
			ActionTitle:   actionTitle,
			ActionEnabled: actionEnabled,
		})
	}
	return cards
}

func desktopSessionsSummary(sessions []desktopSession) string {
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

	if runningCount == 0 {
		return fmt.Sprintf("You don't have any desktop sessions currently running %s", nonRunningString)
	}
	if runningCount == 1 {
		return fmt.Sprintf("You have 1 desktop session currently running %s", nonRunningString)
	}
	return fmt.Sprintf("You have %d desktop sessions currently running %s", runningCount, nonRunningString)
}

func desktopListCommand(ctx context.Context, username string) ([]desktopSession, error) {
	cmd, err := buildDesktopCommand(ctx, username, "list", "--format", "json")
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() != 0 {
			return nil, fmt.Errorf("listing desktop sessions: %s", stderr.String())
		}
		return nil, fmt.Errorf("listing desktop sessions: %w", err)
	}

	var sessions []desktopSession
	if err := json.Unmarshal(stdout.Bytes(), &sessions); err != nil {
		return nil, fmt.Errorf("decoding desktop sessions: %w", err)
	}
	return sessions, nil
}

type terminationResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error"`
	Reason      string `json:"reason"`
}

func desktopKillCommand(ctx context.Context, username, sessionName string) (terminationResponse, error) {
	cmd, err := buildDesktopCommand(ctx, username, "kill", sessionName, "--format", "json")
	if err != nil {
		return terminationResponse{}, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	var response terminationResponse
	if decodeErr := json.Unmarshal(stdout.Bytes(), &response); decodeErr == nil {
		return response, nil
	}

	if runErr != nil {
		if stderr.Len() != 0 {
			return terminationResponse{}, fmt.Errorf("terminating desktop session: %s", stderr.String())
		}
		return terminationResponse{}, fmt.Errorf("terminating desktop session: %w", runErr)
	}
	return terminationResponse{}, fmt.Errorf("decoding desktop termination response: %s", stdout.String())
}

func buildDesktopCommand(ctx context.Context, username string, args ...string) (*exec.Cmd, error) {
	userInfo, err := user.Lookup(username)
	if err != nil {
		return nil, fmt.Errorf("looking up user %q: %w", username, err)
	}

	uid, err := strconv.Atoi(userInfo.Uid)
	if err != nil {
		return nil, fmt.Errorf("parsing uid for user %q: %w", userInfo.Username, err)
	}
	gid, err := strconv.Atoi(userInfo.Gid)
	if err != nil {
		return nil, fmt.Errorf("parsing gid for user %q: %w", userInfo.Username, err)
	}

	groupIDs, err := userInfo.GroupIds()
	if err != nil {
		return nil, fmt.Errorf("looking up groups for user %q: %w", userInfo.Username, err)
	}
	groups := make([]uint32, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		parsed, err := strconv.Atoi(groupID)
		if err != nil {
			return nil, fmt.Errorf("parsing group id %q for user %q: %w", groupID, userInfo.Username, err)
		}
		groups = append(groups, uint32(parsed))
	}

	cmd := exec.CommandContext(ctx, desktopToolPath(), args...)
	cmd.Dir = userInfo.HomeDir
	cmd.Env = desktopCommandEnv(userInfo)
	if os.Geteuid() == 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid:    uint32(uid),
				Gid:    uint32(gid),
				Groups: groups,
			},
		}
	} else if os.Geteuid() != uid {
		return nil, fmt.Errorf("cannot run desktop listing as %q without root privileges", userInfo.Username)
	}
	return cmd, nil
}

func desktopToolPath() string {
	return filepath.Join(env.FlightRoot, "usr", "lib", "flight-core", "flight-desktop")
}

func desktopCommandEnv(userInfo *user.User) []string {
	env := slices.Clone(os.Environ())
	env = slices.DeleteFunc(env, func(envar string) bool {
		parts := strings.SplitN(envar, "=", 2)
		name := parts[0]
		return strings.HasPrefix(name, "XDG_")
	})
	env = upsertEnv(env, "HOME", userInfo.HomeDir)
	env = upsertEnv(env, "LOGNAME", userInfo.Username)
	env = upsertEnv(env, "USER", userInfo.Username)
	return env
}

func upsertEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}
