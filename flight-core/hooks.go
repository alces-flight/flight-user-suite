package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"charm.land/log/v2"
	"github.com/concertim/flight-user-suite/flight/howto_guides"
	"github.com/concertim/flight-user-suite/flight/pkg"
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
	event := cmd.String("event")
	var hooks []*Hook
	var err error
	if event == "" {
		hooks, err = getAllHooks(onlyEnabled)
	} else {
		hooks, err = getEventHooks(event, onlyEnabled)
	}
	if err != nil {
		return err
	}
	if len(hooks) == 0 {
		if onlyEnabled || event != "" {
			fmt.Println("No hooks match the given filters.")
		} else {
			fmt.Println("No hooks found.")
		}
		return nil
	}
	return hooksTable(hooks)
}

func hooksTable(hooks []*Hook) error {
	namecolWidth := 8
	eventcolWidth := 8

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(pkg.AlcesBlue)).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			switch {
			case row == table.HeaderRow:
				return pkg.TableHeaderStyle
			case row%2 == 0:
				style = pkg.TableEvenRowStyle
			default:
				style = pkg.TableOddRowStyle
			}
			switch col {
			case 0:
				return style.MaxWidth(eventcolWidth)
			case 1:
				return style.Width(namecolWidth)
			case 2:
				return style.Width(13)
			}
			return style
		})
	t.Headers("Event", "Name", "Enabled")
	for _, h := range hooks {
		eventcolWidth = max(eventcolWidth, len(h.Event)+2)
		namecolWidth = max(namecolWidth, len(h.Name)+2)
		enabledText := "\u274c Disabled"
		if h.Enabled {
			enabledText = "\u2705 Enabled"
		}
		t.Row(h.Event, h.Name, enabledText)
	}
	_, err := lipgloss.Println(t)
	return err
}

func getAllHooks(onlyEnabled bool) ([]*Hook, error) {
	hooks := make([]*Hook, 0)
	for _, event := range validEvents {
		hs, err := getEventHooks(event, onlyEnabled)
		if err != nil {
			return hooks, err
		}
		hooks = append(hooks, hs...)
	}
	return hooks, nil
}

func getEventHooks(event string, onlyEnabled bool) ([]*Hook, error) {
	hookDir := filepath.Join(hookDir, event)
	log.Debug("getting hooks", "dir", hookDir, "onlyEnabled", onlyEnabled)
	entries, err := os.ReadDir(hookDir)
	if err != nil {
		return nil, fmt.Errorf("listing hooks: %w", err)
	}
	hooks := make([]*Hook, 0)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("reading hook info: %w", err)
		}
		enabled := info.Mode()&0111 != 0
		hook := &Hook{Name: entry.Name(), Event: event, Enabled: enabled}
		if !onlyEnabled || hook.Enabled {
			hooks = append(hooks, hook)
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
	if err := howto_guides.CreateHowtoSymlinks(hook, false); err != nil {
		log.Debug("Error installing howtos", "hook", hook, "err", err)
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
	if err := howto_guides.RemoveHowtoSymlinks(hook, false); err != nil {
		return fmt.Errorf("removing howto symlinks: %w", err)
	}
	log.Printf("Disabled flight %s hook", hook)
	return nil
}

type Hook struct {
	Enabled bool
	Event   string
	Name    string
}

type UnknownHook struct {
	Event string
	Hook  string
}

func (ut UnknownHook) Error() string {
	return fmt.Sprintf("Unknown %s hook: %s", ut.Event, ut.Hook)
}
