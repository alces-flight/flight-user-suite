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
	"strings"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

func toolPath(tool string) string {
	return filepath.Join(toolDir, fmt.Sprintf("flight-%s", tool))
}

func howTosDir(tool string) string {
	return filepath.Join(flightRoot, "usr", "share", "doc", fmt.Sprintf("flight-%s", tool))
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
	for _, tool := range tools {
		fmt.Println(tool)
	}
	return nil
}

func getTools(onlyEnabled bool) ([]string, error) {
	log.Debug("getting tools", "dir", toolDir, "onlyEnabled", onlyEnabled)
	entries, err := os.ReadDir(toolDir)
	if err != nil {
		return nil, fmt.Errorf("listing tools: %w", err)
	}
	tools := make([]string, 0)
	for _, entry := range entries {
		if tool, hasPrefix := strings.CutPrefix(entry.Name(), "flight-"); hasPrefix {
			if onlyEnabled {
				info, err := entry.Info()
				if err != nil {
					return nil, fmt.Errorf("reading tool info: %w", err)
				}
				if info.Mode()&0111 != 0 {
					tools = append(tools, tool)
				}
			} else {
				tools = append(tools, tool)
			}
		}
	}
	return tools, nil
}

func enableTool(ctx context.Context, cmd *cli.Command) error {
	tool := cmd.StringArg("tool")
	tp := toolPath(tool)
	log.Debug("Enabling", "tool", tool, "path", tp)
	if err := os.Chmod(tp, 0555); err != nil {
		return transformToolError(tool, err)
	}
	createHowtoSymlinks(tool)
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
	if err := removeHowtoSymlinks(tool); err != nil {
		return fmt.Errorf("removing howto symlinks: %w", err)
	}
	log.Printf("Disabled flight %s tool", tool)
	return nil
}

func createHowtoSymlinks(tool string) error {
	tgtDir := filepath.Join(flightRoot, "usr", "share", "doc", "howtos-enabled")
	srcDir := howTosDir(tool)
	matches, err := filepath.Glob(filepath.Join(srcDir, "*.md"))
	if err != nil {
		return fmt.Errorf("globbing howtos: %w", err)
	}
	for _, oldpath := range matches {
		newpath := filepath.Join(tgtDir, filepath.Base(oldpath))
		log.Debug("Creating howto symlink", "target", oldpath, "link_name", newpath)
		if err = os.Symlink(oldpath, newpath); err != nil {
			return fmt.Errorf("creating howto symlink: %w", err)
		}
	}
	return nil
}

func removeHowtoSymlinks(tool string) error {
	symDir := filepath.Join(flightRoot, "usr", "share", "doc", "howtos-enabled")
	srcDir := howTosDir(tool)

	entries, err := os.ReadDir(symDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			symTgt, err := os.Readlink(filepath.Join(symDir, entry.Name()))
			if err != nil {
				return err
			}
			if !filepath.IsAbs(symTgt) {
				symTgt = filepath.Join(symDir, symTgt)
			}
			symTgtDir := filepath.Dir(filepath.Clean(symTgt))
			if symTgtDir == srcDir {
				log.Debug("Removing howto symlink", "target", symTgt, "link_name", filepath.Join(symDir, entry.Name()))
				os.Remove(filepath.Join(symDir, entry.Name()))
			}
		}
	}
	return nil
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
