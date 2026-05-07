package userroles

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"charm.land/log/v2"
	"github.com/urfave/cli/v3"
)

// IsAdmin returns true if the user running the process is considered an admin
// and false otherwise.
//
// "Considered an admin" is defined as: has password-less `sudo` access to run
// the same `flight` executable that AsRoot will re-exec.
func IsAdmin() bool {
	if os.Geteuid() == 0 {
		return true
	}
	path, err := flightCommandPath()
	if err != nil {
		return false
	}
	exe := exec.Command("sudo", "-ln", "--", path)
	if err := exe.Run(); err != nil {
		return false
	}
	return true
}

func flightCommandPath() (string, error) {
	path, err := os.Executable()
	if err == nil {
		path, evalErr := filepath.EvalSymlinks(path)
		if evalErr == nil {
			return path, nil
		}
		return path, nil
	}

	// Fall back to resolving argv[0] for environments where os.Executable
	// cannot determine the path.
	if filepath.IsAbs(os.Args[0]) {
		return os.Args[0], nil
	}

	path, err = exec.LookPath(os.Args[0])
	if errors.Is(err, exec.ErrDot) {
		dir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolving flight command path: %w", err)
		}
		return filepath.Join(dir, path), nil
	}
	if err != nil {
		return "", fmt.Errorf("resolving flight command path: %w", err)
	}
	return path, nil
}

func AsRoot(wrapped cli.ActionFunc) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		if os.Geteuid() == 0 {
			// We're running as root, just run the wrapped function.
			return wrapped(ctx, cmd)
		}

		if !IsAdmin() {
			// This should not have happened.
			return fmt.Errorf("command '%s' not available to non-admin users", cmd.Name)
		}

		path, err := flightCommandPath()
		if err != nil {
			// This should not have happened.
			return fmt.Errorf("unexpected error trying to run as root: %w", err)
		}

		// Let's re-exec with sudo, preserving just enough of the environment
		// to function.
		args := []string{"--preserve-env=FLIGHT_ROOT,FLIGHT_STATE_ROOT", path}
		args = append(args, os.Args[1:len(os.Args)]...)

		exe := exec.CommandContext(ctx, "sudo", args...)
		exe.Env = slices.Clone(os.Environ())
		exe.Stdout = os.Stdout
		exe.Stderr = os.Stderr
		exe.Stdin = os.Stdin
		log.Debug("Re-execing as root", "cmd", exe.String())
		return exe.Run()
	}
}
