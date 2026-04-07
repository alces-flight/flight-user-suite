package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"charm.land/log/v2"
	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

func envWhitelist() []string {
	config, err := loadConfig()
	if err != nil {
		whitelist := []string{"PWD", "HOME", "LANG", "USER", "UID", "PATH", "VNCDESKTOP", "DISPLAY", "FLIGHT_ROOT"}
		log.Debug("Loading config failed: using default environment whitelist", "err", err)
		return whitelist
	}
	whitelist := make([]string, 0, len(config.EnvWhitelist))
	for _, item := range config.EnvWhitelist {
		whitelist = append(whitelist, strings.TrimSpace(item))
	}
	return whitelist
}

type sessionState string

var (
	New    sessionState = "new"
	Active sessionState = "active"
	Broken sessionState = "broken"
	Exited sessionState = "exited"
)

type Session struct {
	ID           string          `yaml:"id"`
	Name         string          `yaml:"name"`
	SessionType  string          `yaml:"session_type"`
	Password     string          `yaml:"password"`
	SessionState sessionState    `yaml:"state"`
	Geometry     string          `yaml:"geometry"`
	CreatedAt    time.Time       `yaml:"created_at"`
	Metadata     sessionMetadata `yaml:"metadata"`
}

func loadAllSessions() ([]*Session, error) {
	glob := filepath.Join(xdg.StateHome, "flight", "desktop", "sessions", "*", "metadata.yml")
	log.Debug("Loading all sessions", "glob", glob)
	sessions := make([]*Session, 0)
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, fmt.Errorf("loading sessions: %w", err)
	}
	for _, match := range matches {
		id := filepath.Base(filepath.Dir(match))
		session, err := loadSession(id)
		if err != nil {
			log.Debug("Skipping bad session", "match", match, "err", err)
			continue
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func loadSession(id string) (*Session, error) {
	session := &Session{ID: id}
	log.Debug("Loading session", "sessionDir", session.sessionDir())
	info, err := os.Stat(session.sessionDir())
	if err != nil {
		log.Debug("Error checking session dir", "sessionDir", session.sessionDir(), "err", err)
		session.SessionState = Broken
		return session, UnknownSession{Session: id}
	}
	if !info.IsDir() {
		log.Debug("Session dir is not a directory", "sessionDir", session.sessionDir())
		session.SessionState = Broken
		return session, UnknownSession{Session: id}
	}

	data, err := os.ReadFile(session.metadataFile())
	if err != nil {
		log.Debug("Reading session metadata", "metadataFile", session.metadataFile(), "err", err)
		session.SessionState = Broken
		return session, nil
	}
	err = yaml.Unmarshal(data, &session)
	if err != nil {
		log.Debug("Loading session metadata", "metadataFile", session.metadataFile(), "err", err)
		session.SessionState = Broken
		return session, nil
	}
	if !session.isActive() {
		session.SessionState = Exited
	}
	return session, nil
}

func (s *Session) Start(ctx context.Context) error {
	s.CreatedAt = time.Now()
	if err := s.mkSessionDir(); err != nil {
		return err
	}
	if err := s.createPassword(); err != nil {
		return err
	}
	if err := s.installSessionScript(); err != nil {
		return err
	}
	if err := s.startVNC(ctx, xdg.Home); err != nil {
		s.SessionState = Broken
		saveErr := s.Save()
		if saveErr != nil {
			log.Debug("Failed to save failed session", "save.err", saveErr, "err", err)
		}
		return fmt.Errorf("staring VNC server: %w", err)
	}
	s.SessionState = Active
	err := s.Save()
	if err != nil {
		return fmt.Errorf("saving session: %w", err)
	}
	return nil
}

func (s *Session) Save() error {
	data, err := yaml.Marshal(&s)
	if err != nil {
		return fmt.Errorf("saving session: %w", err)
	}
	metadataFile := s.metadataFile()
	log.Debug("saving session", "file", metadataFile)
	os.WriteFile(metadataFile, data, 0o600)
	return nil
}

func (s *Session) Kill(ctx context.Context) error {
	args := []string{
		"-kill",
		"-sessiondir", s.sessionDir(),
	}
	cmd := exec.CommandContext(ctx, libexecPath("vncserver"), args...)
	output, err := cmd.CombinedOutput()
	if exitError, ok := errors.AsType[*exec.ExitError](err); ok {
		log.Debug("vncserver", "stdout/stderr", string(output))
		return SilentExitError{ExitCode: exitError.ExitCode(), exitError: exitError}
	} else if err != nil {
		log.Debug("vncserver", "stdout/stderr", string(output))
		return fmt.Errorf("killing VNC server: %w", err)
	}
	return s.RemoveSessionDir()
}

func (s *Session) RemoveSessionDir() error {
	err := os.RemoveAll(s.sessionDir())
	if err != nil {
		return fmt.Errorf("removing session dir: %w", err)
	}
	return nil
}

func (s Session) PrimaryIP() netip.Addr {
	ip, err := getPrimaryIP()
	if err != nil {
		log.Debug("unable to get primary IP", "err", err)
	}
	return ip
}

func (s Session) Port() int {
	display, err := strconv.Atoi(s.Display())
	if err != nil {
		log.Debug("unable to parse display", "display", s.Display(), "err", err)
		return -1
	}
	return 5900 + display
}

func (s Session) PrimaryConnectionString() string {
	if s.SessionState == Broken {
		return ""
	}
	ip := s.PrimaryIP().String()
	return fmt.Sprintf("%s:%d", ip, s.Port())
}

func (s Session) Display() string {
	return s.Metadata.Display
}

func (s *Session) mkSessionDir() error {
	dir := s.sessionDir()
	log.Debug("creating session dir", "dir", dir)
	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		return fmt.Errorf("creating session directory: %w", err)
	}
	return nil
}

func (s *Session) sessionDir() string {
	return filepath.Join(xdg.StateHome, "flight", "desktop", "sessions", s.ID)
}

func (s *Session) createPassword() error {
	log.Debug("creating password", "file", s.passwordFile())
	// TODO: Support alternative password generation, e.g., apg.
	s.Password = rand.Text()[0:8]
	vncpasswd := "/usr/bin/vncpasswd"
	cmd := exec.Command(vncpasswd, "-f")
	cmd.Stdin = bytes.NewReader([]byte(s.Password))
	output, err := cmd.Output()
	if err != nil {
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
			log.Debug("vncpasswd output", "stdout", string(output), "stderr", string(ee.Stderr))
		}
		return fmt.Errorf("setting password: %w", err)
	}
	err = os.WriteFile(s.passwordFile(), output, 0o400)
	if err != nil {
		return fmt.Errorf("saving password: %w", err)
	}
	return nil
}

func (s *Session) installSessionScript() error {
	srcPath := filepath.Join(flightRoot, "usr", "lib", "desktop", "types", s.SessionType, "session.sh")
	dstPath := s.sessionScript()
	log.Debug("installing session script", "src", srcPath, "dst", dstPath)
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("installing session script: %w", err)
	}
	defer src.Close()
	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	if err != nil {
		return fmt.Errorf("installing session script: %w", err)
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("installing session script: %w", err)
	}
	return nil
}

// Return a copy of the current environment with only whitelisted items remaining.
func (s *Session) cleanEnvironment() []string {
	whitelist := envWhitelist()
	log.Debug("Sanitising environment", "whitelist", whitelist)
	clean := make([]string, 0, len(os.Environ()))
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		key := parts[0]
		if slices.Contains(whitelist, key) {
			clean = append(clean, kv)
		}
	}
	return clean
}

func (s *Session) passwordFile() string {
	return filepath.Join(s.sessionDir(), "password.dat")
}

func (s *Session) sessionScript() string {
	return filepath.Join(s.sessionDir(), "session.sh")
}

func (s *Session) metadataFile() string {
	return filepath.Join(s.sessionDir(), "metadata.yml")
}

func (s *Session) isActive() bool {
	b, err := os.ReadFile(s.Metadata.Pidfile)
	if err != nil {
		log.Debug("Unable to read", "pidfile", s.Metadata.Pidfile, "err", err)
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		log.Debug("Unable to parse", "pidfile", s.Metadata.Pidfile, "err", err)
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		log.Debug("Unable to find process", "pid", pid, "err", err)
		return false
	}
	err = p.Signal(syscall.Signal(0))
	return err == nil
}

func (s *Session) startVNC(ctx context.Context, dir string) error {
	passwdFile := s.passwordFile()
	args := []string{
		"-autokill",
		"-sessiondir", s.sessionDir(),
		"-sessionscript", s.sessionScript(),
		"-vncpasswd", passwdFile,
		"-exedir", "/usr/bin",
		"-geometry", s.Geometry,
	}
	cmd := exec.CommandContext(ctx, libexecPath("vncserver"), args...)
	cmd.Dir = dir
	cmd.Env = s.cleanEnvironment()

	output, err := cmd.Output()
	if err != nil {
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
			log.Debug("vncserver", "stdout", string(output), "stderr", string(ee.Stderr))
		}
		s.parseVNCOutput(output) // nolint:errcheck
		return err
	}
	return s.parseVNCOutput(output)
}

func (s *Session) parseVNCOutput(output []byte) error {
	inYamlBlock := false
	var yamlString strings.Builder
	for line := range strings.Lines(string(output)) {
		if line == "<YAML>\n" {
			inYamlBlock = true
		} else if line == "</YAML>\n" {
			inYamlBlock = false
		} else if inYamlBlock {
			yamlString.WriteString(line)
		}
	}
	ys := yamlString.String()
	var md sessionMetadata
	err := yaml.Unmarshal([]byte(ys), &md)
	if err != nil {
		return err
	}
	s.Metadata = md
	return nil
}

type sessionMetadata struct {
	Host    string `yaml:"host"`
	Display string `yaml:"display"`
	Log     string `yaml:"log"`
	Pidfile string `yaml:"pidfile"`
}

// Wrapper around exec.ExitError that avoids the default handling by urfave/cli.
type SilentExitError struct {
	ExitCode  int
	exitError error
}

func (ee SilentExitError) Error() string {
	return ee.exitError.Error()
}
