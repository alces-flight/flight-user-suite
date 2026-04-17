package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"
)

// TODO: Replace these functions with [pkg.WritePidfile] and [pkg.ReadPidfile]
// when the issue with air/replace directives/our go module names is resolved.

func writePidfile(pidfile string, pid int) error {
	existingPID, err := readPidfile(pidfile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if existingPID != 0 {
		return fmt.Errorf("process with PID %d is still running", existingPID)
	}
	return os.WriteFile(pidfile, []byte(strconv.Itoa(pid)), 0o644)
}

func readPidfile(pidfile string) (int, error) {
	data, err := os.ReadFile(pidfile)
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
