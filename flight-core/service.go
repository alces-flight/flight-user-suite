package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"charm.land/log/v2"
	"github.com/concertim/flight-user-suite/flight/pidfile"
)

type Service struct {
	ID   string
	Name string
}

func (s *Service) ExePath() string {
	return filepath.Join(env.FlightRoot, "usr", "libexec", s.ID, "service")
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
	pid, err := pidfile.Read(s.PidfilePath())
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
	pid, _ := pidfile.Read(s.PidfilePath())
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
