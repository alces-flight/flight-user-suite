package userroles

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"

	"charm.land/log/v2"
	"github.com/urfave/cli/v3"
)

// IsAdmin returns true if the user running the process is considered an admin
// and false otherwise.
//
// Considered an admin is defined as: has password-less sudo access.
func IsAdmin() bool {
	if os.Geteuid() == 0 {
		return true
	}
	exe := exec.Command("sudo", "-ln")
	if err := exe.Run(); err != nil {
		return false
	}
	return true
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

		// Let's re-exec with sudo, preserving just enough of the environment
		// to function.
		args := []string{"--preserve-env=FLIGHT_ROOT,FLIGHT_STATE_ROOT"}
		args = append(args, os.Args...)

		exe := exec.CommandContext(ctx, "sudo", args...)
		exe.Env = slices.Clone(os.Environ())
		exe.Stdout = os.Stdout
		exe.Stderr = os.Stderr
		exe.Stdin = os.Stdin
		log.Debug("Re-execing as root", "cmd", exe.String())
		return exe.Run()
	}
}
