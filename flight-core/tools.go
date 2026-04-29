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

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"charm.land/log/v2"
	"github.com/concertim/flight-user-suite/flight/cliui"
	"github.com/concertim/flight-user-suite/flight/toolset"
	"github.com/urfave/cli/v3"
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
	tools, err := toolset.GetTools(env.FlightRoot, onlyEnabled)
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

func toolsTable(tools []*toolset.Tool) error {
	namecolWidth := 8

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(cliui.AlcesBlue)).
		StyleFunc(func(row, col int) lipgloss.Style {
			var style lipgloss.Style
			switch {
			case row == table.HeaderRow:
				return cliui.TableHeaderStyle
			case row%2 == 0:
				style = cliui.TableEvenRowStyle
			default:
				style = cliui.TableOddRowStyle
			}
			switch col {
			case 0:
				return style.Width(namecolWidth)
			case 2:
				return style.Width(13)
			}
			return style
		})
	t.Headers("Name", "Description", "Enabled")
	for _, tool := range tools {
		namecolWidth = max(namecolWidth, len(tool.Name)+2)
		enabledText := "\u274c Disabled"
		if tool.Enabled {
			enabledText = "\u2705 Enabled"
		}
		t.Row(tool.Name, tool.Synopsis, enabledText)
	}
	_, err := lipgloss.Println(t)
	return err
}

func enableTool(ctx context.Context, cmd *cli.Command) error {
	tool := cmd.StringArg("tool")
	tp := toolPath(tool)
	log.Debug("Enabling", "tool", tool, "path", tp)
	if err := os.Chmod(tp, 0555); err != nil {
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

func runToolAction(tool *toolset.Tool) func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		return runTool(ctx, tool, cmd.Args().Slice())
	}
}

func buildToolExecCmd(ctx context.Context, tool *toolset.Tool, args []string) *exec.Cmd {
	tp := toolPath(tool.Name)
	log.Debug("Execing", "tool", tool.Name, "path", tp, "args", args)
	exe := exec.CommandContext(ctx, tp, args...)
	exe.Env = slices.Clone(os.Environ())
	exe.Env = append(exe.Env, fmt.Sprintf("FLIGHT_PROGRAM_NAME=flight %s", tool.Name))
	return exe
}

func runTool(ctx context.Context, tool *toolset.Tool, args []string) error {
	exe := buildToolExecCmd(ctx, tool, args)
	exe.Stdout = os.Stdout
	exe.Stderr = os.Stderr
	exe.Stdin = os.Stdin
	err := exe.Run()
	return transformToolError(tool.Name, err)
}

// Run the specified tool and return its Stdout.
func runToolWithOutput(ctx context.Context, tool *toolset.Tool, args []string) ([]byte, error) {
	exe := buildToolExecCmd(ctx, tool, args)
	output, err := exe.Output()
	return output, transformToolError(tool.Name, err)
}

type DisabledTool struct {
	Tool string
}

func (dt DisabledTool) Error() string {
	return fmt.Sprintf("The %s tool is not enabled", dt.Tool)
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
