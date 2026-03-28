package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"charm.land/log/v2"
	"github.com/urfave/cli/v3"
)

func hookPath(event, hook string) string {
	return filepath.Join(hookDir, event, hook)
}

func transformHookError(event, hook string, err error) error {
	if pathError, ok := errors.AsType[*fs.PathError](err); ok {
		if pathError.Err.Error() == "no such file or directory" {
			return UnknownHook{Event: event, Hook: hook}
		}
	}
	return err
}

func listHooks(_ context.Context, cmd *cli.Command) error {
	onlyEnabled := cmd.Bool("enabled")
	event := cmd.StringArg("event")
	hooks, err := getHooks(event, onlyEnabled)
	if err != nil {
		return err
	}
	for _, hook := range hooks {
		fmt.Println(hook)
	}
	return nil
}

func getHooks(event string, onlyEnabled bool) ([]string, error) {
	if event != "" {
		hookDir = filepath.Join(hookDir, event)
	}
	log.Debug("getting hooks", "dir", hookDir, "onlyEnabled", onlyEnabled)
	entries, err := os.ReadDir(hookDir)
	if err != nil {
		return nil, fmt.Errorf("listing hooks: %w", err)
	}
	hooks := make([]string, 0)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if onlyEnabled {
			info, err := entry.Info()
			if err != nil {
				return nil, fmt.Errorf("reading hook info: %w", err)
			}
			if info.Mode()&0111 != 0 {
				hooks = append(hooks, entry.Name())
			}
		} else {
			hooks = append(hooks, entry.Name())
		}
	}
	return hooks, nil
}

func enableHook(ctx context.Context, cmd *cli.Command) error {
	event := cmd.StringArg("event")
	hook := cmd.StringArg("hook")
	hp := hookPath(event, hook)
	log.Debug("Enabling", "hook", hook, "path", hp)
	if strings.HasPrefix(hook, ".") {
		return UnknownHook{Event: event, Hook: hook}
	}
	if err := os.Chmod(hp, 0755); err != nil {
		return transformHookError(event, hook, err)
	}
	log.Printf("Enabled %s hook", hook)
	return nil
}

func disableHook(ctx context.Context, cmd *cli.Command) error {
	event := cmd.StringArg("event")
	hook := cmd.StringArg("hook")
	hp := hookPath(event, hook)
	log.Debug("Disabling", "hook", hook, "path", hp)
	if strings.HasPrefix(hook, ".") {
		return UnknownHook{Event: event, Hook: hook}
	}
	if err := os.Chmod(hp, 0444); err != nil {
		return transformHookError(event, hook, err)
	}
	log.Printf("Disabled flight %s hook", hook)
	return nil
}

type UnknownHook struct {
	Event string
	Hook  string
}

func (ut UnknownHook) Error() string {
	return fmt.Sprintf("Unknown %s hook: %s", ut.Event, ut.Hook)
}
