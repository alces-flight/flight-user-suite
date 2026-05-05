package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"charm.land/log/v2"
	"github.com/concertim/flight-user-suite/flight/pidfile"
	"github.com/concertim/flight-user-suite/flight/process"
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

func (s *Service) Start(ctx context.Context) (*process.Response, error) {
	err := s.mkPidfileDir()
	if err != nil {
		return nil, fmt.Errorf("creating pidfile directory: %w", err)
	}

	pidfilePath := s.PidfilePath()

	extantProcess, _ := pidfile.Read(pidfilePath)

	if extantProcess != 0 {
		return nil, fmt.Errorf("Service %s is already running (PID %d)", s.Name, extantProcess)
	}

	args := []string{"--pidfile", pidfilePath}
	log.Debug("Starting", "service", s.ID, "path", s.ExePath(), "args", args)
	execCmd := exec.CommandContext(ctx, s.ExePath(), args...)
	execCmd.Dir = "/"

	logfilePath := filepath.Join(env.FlightRoot, "var", "log", fmt.Sprintf("%s.log", s.ID))
	err = os.MkdirAll(filepath.Dir(logfilePath), 0o755)
	if err != nil {
		return nil, fmt.Errorf("creating log director: %w", err)
	}
	logfile, err := os.OpenFile(logfilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}
	defer logfile.Close()
	execCmd.Stdout = logfile
	execCmd.Stderr = logfile

	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("os.Pipe() failed: %w", err)
	}
	defer pipeRead.Close()
	defer pipeWrite.Close()
	execCmd.ExtraFiles = []*os.File{pipeWrite}

	// TODO: What environment do we want to run in? What did flight-service do?
	// cmd.Env = s.cleanEnvironment()
	startErr := execCmd.Start()

	if startErr != nil {
		return nil, startErr
	}

	pipeWrite.Close()

	reader := bufio.NewReader(pipeRead)

	var response process.Response
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("Failed to read response from service process: %w", err)
	}

	return &response, nil
}

func (s *Service) Kill() error {
	log.Debug("Killing service process", "pidfile", s.PidfilePath(), "name", s.ID)
	pid, err := pidfile.Read(s.PidfilePath())
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		return err
	}
	if pid == 0 {
		return errors.New("No running process found")
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
