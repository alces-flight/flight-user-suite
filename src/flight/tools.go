package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

func toolPath(tool string) string {
	return filepath.Join(flightRoot, "usr", "lib", "flight-core", fmt.Sprintf("flight-%s", tool))
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

func enableTool(ctx context.Context, cmd *cli.Command) error {
	tool := cmd.StringArg("tool")
	tp := toolPath(tool)
	log.Debug("Enabling", "tool", tool, "path", tp)
	err := os.Chmod(tp, 0555)
	if err == nil {
		log.Printf("Enabled flight %s tool", tool)
		return nil
	}
	return transformToolError(tool, err)
}

func disableTool(ctx context.Context, cmd *cli.Command) error {
	tool := cmd.StringArg("tool")
	tp := toolPath(tool)
	log.Debug("Disabling", "tool", tool, "path", tp)
	err := os.Chmod(tp, 0444)
	if err == nil {
		log.Printf("Disabled flight %s tool", tool)
		return nil
	}
	return transformToolError(tool, err)
}

func runTool(tool string) func(ctx context.Context, cmd *cli.Command) error {
	run := func(ctx context.Context, cmd *cli.Command) error {
		tp := toolPath(tool)
		log.Debug("Execing", "tool", tool, "path", tp, "args", cmd.Args().Slice())

		exe := exec.CommandContext(ctx, tp, cmd.Args().Slice()...)
		stdout, err := exe.StdoutPipe()
		if err != nil {
			return fmt.Errorf("creating stdout pipe: %w", err)
		}
		stderr, err := exe.StderrPipe()
		if err != nil {
			return fmt.Errorf("creating stderr pipe: %w", err)
		}

		go func() { io.Copy(os.Stdout, stdout) }()
		go func() { io.Copy(os.Stderr, stderr) }()

		err = exe.Run()
		return transformToolError(tool, err)
	}

	return run
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
