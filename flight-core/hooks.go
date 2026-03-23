package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

func hookPath(hook string) string {
	return filepath.Join(hookDir, hook)
}

func transformHookError(hook string, err error) error {
	if pathError, ok := errors.AsType[*fs.PathError](err); ok {
		if pathError.Err.Error() == "no such file or directory" {
			return UnknownHook{Hook: hook}
		}
	}
	return err
}

func listHooks(_ context.Context, cmd *cli.Command) error {
	onlyEnabled := cmd.Bool("enabled")
	hooks, err := getHooks(onlyEnabled)
	if err != nil {
		return err
	}
	for _, hook := range hooks {
		fmt.Println(hook)
	}
	return nil
}

func getHooks(onlyEnabled bool) ([]string, error) {
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
	hook := cmd.StringArg("hook")
	hp := hookPath(hook)
	log.Debug("Enabling", "hook", hook, "path", hp)
	if strings.HasPrefix(hook, ".") {
		return UnknownHook{Hook: hook}
	}
	if err := os.Chmod(hp, 0755); err != nil {
		return transformHookError(hook, err)
	}
	log.Printf("Enabled %s hook", hook)
	return nil
}

func disableHook(ctx context.Context, cmd *cli.Command) error {
	hook := cmd.StringArg("hook")
	hp := hookPath(hook)
	log.Debug("Disabling", "hook", hook, "path", hp)
	if strings.HasPrefix(hook, ".") {
		return UnknownHook{Hook: hook}
	}
	if err := os.Chmod(hp, 0444); err != nil {
		return transformHookError(hook, err)
	}
	log.Printf("Disabled flight %s hook", hook)
	return nil
}

type UnknownHook struct {
	Hook string
}

func (ut UnknownHook) Error() string {
	return fmt.Sprintf("Unknown hook: %s", ut.Hook)
}
