package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"charm.land/log/v2"
)

type Service struct {
	ID   string
	Name string
}

func (s *Service) ExePath() string {
	return filepath.Join(flightRoot, "usr", "libexec", s.ID, "service")
}

func (s *Service) PidfilePath() string {
	return filepath.Join("/", "var", "run", "flight", fmt.Sprintf("%s.pid", s.ID))
}

func (s *Service) Start(ctx context.Context) error {
	err := s.mkPidfileDir()
	if err != nil {
		return fmt.Errorf("creating pidfile directory: %w", err)
	}

	args := []string{"--pidfile", s.PidfilePath()}
	log.Debug("Starting", "service", s.ID, "path", s.ExePath(), "args", args)
	execCmd := exec.CommandContext(ctx, s.ExePath(), args...)
	execCmd.Dir = "/"
	// TODO: What environment do we want to run in? What did flight-service do?
	// cmd.Env = s.cleanEnvironment()
	return execCmd.Start()
}

func (s *Service) Kill() error {
	log.Debug("Killing service process", "pidfile", s.PidfilePath(), "name", s.ID)
	pid, err := readPidfile(s.PidfilePath())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if pid == 0 {
		// Process is no longer running.
		return nil
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	err = process.Signal(os.Interrupt)
	if err != nil {
		return err
	}
	err = os.Remove(s.PidfilePath())
	if err != nil {
		log.Debug("Error removing pidfile", "pidfile", s.PidfilePath(), "err", err)
	}
	return nil
}

func (s *Service) State() string {
	pid, _ := readPidfile(s.PidfilePath())
	if pid == 0 {
		return "Stopped"
	}
	return "Running"
}

func (s *Service) mkPidfileDir() error {
	dir := filepath.Dir(s.PidfilePath())
	log.Debug("Creating pidfile directory", "path", dir)
	return os.MkdirAll(dir, 0o755)
}

// Read the PID file at path. Return the PID contained in the file if it
// contains a valid PID for a running process.  Otherwise return 0.
func readPidfile(path string) (int, error) {
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
