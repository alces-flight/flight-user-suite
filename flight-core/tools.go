package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"charm.land/log/v2"
	"github.com/concertim/flight-user-suite/flight/pkg"
	"github.com/urfave/cli/v3"
	userpermissions "github.com/wneessen/go-fileperm"
)

func toolPath(tool string) string {
	return filepath.Join(toolDir, fmt.Sprintf("flight-%s", tool))
}

func transformToolError(tool string, err error) error {
	if pathError, ok := errors.AsType[*fs.PathError](err); ok {
		if pathError.Err.Error() == "no such file or directory" {
			return UnknownTool{Tool: tool}
		}
		if pathError.Err.Error() == "permission denied" {
			return DisabledTool{Tool: tool}
		}
	} else if exitError, ok := errors.AsType[*exec.ExitError](err); ok {
		return SilentExitError{ExitCode: exitError.ExitCode(), exitError: exitError}
	}
	return err
}

func listTools(ctx context.Context, cmd *cli.Command) error {
	onlyEnabled := cmd.Bool("enabled")
	tools, err := getTools(onlyEnabled)
	if err != nil {
		return err
	}
	if len(tools) == 0 {
		if onlyEnabled {
			fmt.Println("No tools are enabled.")
		} else {
			fmt.Println("No tools found.")
		}
		return nil
	}
	return toolsTable(tools)
}

func toolsTable(tools []*Tool) error {
	namecolWidth := 8

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
				return style.Width(namecolWidth)
			case 2:
				return style.Width(15)
			}
			return style
		})
	t.Headers("Name", "Description", "Enabled")
	for _, tool := range tools {
		namecolWidth = max(namecolWidth, len(tool.Name)+2)
		enabledText := "\u274c Disabled"
		if tool.Enabled && tool.AdminOnly {
			enabledText = "\u2705 Admin only"
		} else if tool.Enabled {
			enabledText = "\u2705 Enabled"
		}
		t.Row(tool.Name, tool.Synopsis, enabledText)
	}
	_, err := lipgloss.Println(t)
	return err
}

func getTools(onlyEnabled bool) ([]*Tool, error) {
	log.Debug("getting tools", "dir", toolDir, "onlyEnabled", onlyEnabled)

	toolSynopsisDir := filepath.Join(flightRoot, "usr", "share", "doc", "tools")

	entries, err := os.ReadDir(toolDir)
	if err != nil {
		return nil, fmt.Errorf("listing tools: %w", err)
	}
	tools := make([]*Tool, 0)
	for _, entry := range entries {
		if toolName, hasPrefix := strings.CutPrefix(entry.Name(), "flight-"); hasPrefix {
			enabled := isUserExecutable(filepath.Join(toolDir, entry.Name()))
			adminOnly := isAdminOnly(entry)
			synopsisFile := filepath.Join(toolSynopsisDir, entry.Name())
			synopsis, _ := os.ReadFile(synopsisFile)
			tool := &Tool{
				AdminOnly: adminOnly,
				Enabled:   enabled,
				Name:      toolName,
				Synopsis:  strings.TrimSpace(string(synopsis)),
			}
			if !onlyEnabled || tool.Enabled {
				tools = append(tools, tool)
			}
		}
	}
	return tools, nil
}

// Return true if the current user can execute the file at the given path.
func isUserExecutable(path string) bool {
	up, err := userpermissions.New(path)
	if err != nil {
		log.Debug("Error checking file permissions", "path", path, "err", err)
		return false
	}
	return up.UserExecutable()
}

// Return true if the file at the given path is executable only for admin
// users.
func isAdminOnly(entry os.DirEntry) bool {
	// We make the assumption here that all tools are owned by root:root, and
	// that a tool is admin only if only the user executable permission bit is
	// set.  This currently matches our build process and the behaviour of
	// [enableTool].
	info, err := entry.Info()
	if err != nil {
		log.Debug("Error checking file permissions", "path", entry, "err", err)
		return false
	}
	enabled := info.Mode()&0111 != 0
	enabledForOthers := info.Mode()&0011 != 0
	return enabled && !enabledForOthers
}

func enableTool(ctx context.Context, cmd *cli.Command) error {
	adminOnly := cmd.Bool("admin-only")
	tool := cmd.StringArg("tool")
	tp := toolPath(tool)
	permissions := os.FileMode(0o555)
	if adminOnly {
		permissions = os.FileMode(0o544)
	}
	log.Debug("Enabling", "tool", tool, "path", tp, "adminOnly", adminOnly, "perms", permissions)
	if err := os.Chmod(tp, permissions); err != nil {
		return transformToolError(tool, err)
	}
	if err := createHowtoSymlinks(tool, true); err != nil {
		log.Debug("Error installing howtos", "tool", tool, "err", err)
	}
	log.Printf("Enabled flight %s tool", tool)
	return nil
}

func disableTool(ctx context.Context, cmd *cli.Command) error {
	tool := cmd.StringArg("tool")
	tp := toolPath(tool)
	log.Debug("Disabling", "tool", tool, "path", tp)
	if err := os.Chmod(tp, 0444); err != nil {
		return transformToolError(tool, err)
	}
	if err := removeHowtoSymlinks(tool, true); err != nil {
		return fmt.Errorf("removing howto symlinks: %w", err)
	}
	log.Printf("Disabled flight %s tool", tool)
	return nil
}

func runToolAction(tool *Tool) func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTool(ctx, tool, cmd.Args().Slice())
	}
}

func buildToolExecCmd(ctx context.Context, tool *Tool, args []string) *exec.Cmd {
	tp := toolPath(tool.Name)
	log.Debug("Execing", "tool", tool.Name, "path", tp, "args", args)
	exe := exec.CommandContext(ctx, tp, args...)
	exe.Env = slices.Clone(os.Environ())
	exe.Env = append(exe.Env, fmt.Sprintf("FLIGHT_PROGRAM_NAME=flight %s", tool.Name))
	return exe
}

func runTool(ctx context.Context, tool *Tool, args []string) error {
	exe := buildToolExecCmd(ctx, tool, args)
	exe.Stdout = os.Stdout
	exe.Stderr = os.Stderr
	err := exe.Run()
	return transformToolError(tool.Name, err)
}

// Run the specified tool and return its Stdout.
func runToolWithOutput(ctx context.Context, tool *Tool, args []string) ([]byte, error) {
	exe := buildToolExecCmd(ctx, tool, args)
	output, err := exe.Output()
	return output, transformToolError(tool.Name, err)
}

type Tool struct {
	AdminOnly bool
	Enabled   bool
	Name      string
	Synopsis  string
}

type DisabledTool struct {
	Tool string
}

func (ut DisabledTool) Error() string {
	return fmt.Sprintf("The %s tool is not enabled", ut.Tool)
}

type UnknownTool struct {
	Tool string
}

func (ut UnknownTool) Error() string {
	return fmt.Sprintf("Unknown tool: %s", ut.Tool)
}

// Wrapper around exec.ExitError that avoids the default handling by urfave/cli.
type SilentExitError struct {
	ExitCode  int
	exitError error
}

func (ee SilentExitError) Error() string {
	return ee.exitError.Error()
}
