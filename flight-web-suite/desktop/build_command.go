// Package desktop provides functions and types for running the
// `flight-desktop` CLI and parsing its output.
package desktop

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

func desktopToolPath(env configenv.Env) string {
	return filepath.Join(env.FlightRoot, "usr", "lib", "flight-core", "flight-desktop")
}

func buildDesktopCommand(ctx context.Context, env configenv.Env, username string, args ...string) (*exec.Cmd, error) {
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

	cmd := exec.CommandContext(ctx, desktopToolPath(env), args...)
	cmd.Dir = userInfo.HomeDir
	cmd.Env = commandEnv(userInfo)
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

func commandEnv(userInfo *user.User) []string {
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
