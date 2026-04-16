package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"
)

// Read the PID file at path. Return the PID contained in the file if it
// contains a valid PID for a running process.  Otherwise return 0.
func ReadPidfile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("reading pidfile: %w", err)
	}
	pid, err := strconv.Atoi(string(bytes.TrimSpace(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid content: %w", err)
	}
	if pid == 0 {
		return 0, fmt.Errorf("invalid content: pid=0")
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, nil
	}
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return 0, nil
	}
	return pid, nil
}

func WritePidfile(path string, pid int) error {
	existingPID, err := ReadPidfile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if existingPID != 0 {
		return fmt.Errorf("process with PID %d is still running", existingPID)
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0o644)
}
