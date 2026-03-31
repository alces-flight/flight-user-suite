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
	"time"

	"charm.land/log/v2"
	"github.com/adrg/xdg"
	"github.com/google/uuid"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

var (
	// TODO: Determine this dynamically by listing the correct directory
	// (opt/flight/usr/lib/desktop/types/).
	validTypes            = []string{"terminal", "gnome"}
	validTypeNames string = strings.Join(validTypes, ", ")
)

func libexecPath(relpath string) string {
	return filepath.Join(flightRoot, "usr", "libexec", "desktop", relpath)
}

func startCommand() *cli.Command {
	return &cli.Command{
		Name:        "start",
		Usage:       "Start an interactive desktop session",
		Description: wordwrap.String("Start a new interactive desktop session and display details about the new session.\n\nAvailable desktop types can be shown using the 'avail' command.", maxTextWidth),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Name the desktop session `NAME` so it can be more easily identified.",
			},
			&cli.StringFlag{
				Name:    "geometry",
				Aliases: []string{"g"},
				Usage:   "Set the desktop geometry to `WIDTHxHEIGHT`.",
				Value:   "1024x768",
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "type", UsageText: "<type>"},
		},
		Before: composeBeforeFuncs(assertArgPresent("type"), assertTypeValid("type", 0)),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			sessionType := cmd.StringArg("type")
			fmt.Printf("Starting a '%s' desktop session:\n\n", sessionType)

			// TODO: Display a spinner.

			session := Session{
				UUID:         uuid.New(),
				SessionState: New,
				SessionType:  sessionType,
				Name:         cmd.String("name"),
				Geometry:     cmd.String("geometry"),
			}
			err := session.start(ctx)

			// TODO: Stop the spinner

			if err != nil {
				fmt.Printf("\u274c Starting session\n\n")
				return fmt.Errorf("starting session: %w", err)
			}
			fmt.Printf("\u2705 Starting session\n\n")
			fmt.Printf("A '%s' desktop session has been started.\n", session.SessionType)
			printSessionDetails(session)
			accessSummary(session)
			return nil
		},
	}
}

func assertTypeValid(argName string, argIndex int) cli.BeforeFunc {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {

		event := cmd.Args().Get(argIndex)
		if !slices.Contains(validTypes, event) {
			return ctx, fmt.Errorf(
				"Incorrect Usage: unknown %s '%s'. Valid values are %s.",
				argName,
				event,
				validTypeNames,
			)
		}
		return ctx, nil
	}
}

type sessionState string

var (
	New    sessionState = "new"
	Active sessionState = "active"
	Broken sessionState = "broken"
)

type Session struct {
	UUID         uuid.UUID `yaml:"uuid"`
	Name         string    `yaml:"name"`
	SessionType  string    `yaml:"session_type"`
	Password     string
	SessionState sessionState    `yaml:"state"`
	Geometry     string          `yaml:"geometry"`
	CreatedAt    time.Time       `yaml:"created_at"`
	Metadata     sessionMetadata `yaml:"metadata"`
}

func (s *Session) Kill(ctx context.Context) error {
	args := []string{
		"-kill",
		"-sessiondir", s.sessionDir(),
	}
	cmd := exec.CommandContext(ctx, libexecPath("vncserver"), args...)
	// TODO: Set environment.
	// cmd.Env = []string{}
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

type sessionMetadata struct {
	Host    string `yaml:"host"`
	Display string `yaml:"display"`
	Log     string `yaml:"log"`
	Pidfile string `yaml:"pidfile"`
}

func (s sessionMetadata) Port() int {
	display, err := strconv.Atoi(s.Display)
	if err != nil {
		log.Debug("unable to parse display", "display", s.Display, "err", err)
		return -1
	}
	return 5900 + display

}

func (s Session) PrimaryIP() netip.Addr {
	ip, err := getPrimaryIP()
	if err != nil {
		log.Debug("unable to get primary IP", "err", err)
	}
	return ip
}

func (s *Session) start(ctx context.Context) error {
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
		return fmt.Errorf("staring VNC server: %w", err)
	}
	s.CreatedAt = time.Now()
	s.SessionState = Active
	err := s.save()
	if err != nil {
		return fmt.Errorf("saving session: %w", err)
	}
	return nil
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
	return filepath.Join(xdg.StateHome, "flight", "desktop", "sessions", s.UUID.String())
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

func (s *Session) passwordFile() string {
	return filepath.Join(s.sessionDir(), "password.dat")
}

func (s *Session) sessionScript() string {
	return filepath.Join(s.sessionDir(), "session.sh")
}

func (s *Session) metadataFile() string {
	return filepath.Join(s.sessionDir(), "metadata.yml")
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
	// TODO: Set environment.
	// cmd.Env = []string{}

	output, err := cmd.Output()
	if err != nil {
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
			log.Debug("vncserver", "stdout", string(output), "stderr", string(ee.Stderr))
		}
		return err
	}

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
	err = yaml.Unmarshal([]byte(ys), &md)
	if err != nil {
		return err
	}
	s.Metadata = md

	return nil
}

func (s *Session) save() error {
	data, err := yaml.Marshal(&s)
	if err != nil {
		return fmt.Errorf("saving session: %w", err)
	}
	metadataFile := s.metadataFile()
	log.Debug("saving session", "file", metadataFile)
	os.WriteFile(metadataFile, data, 0o600)
	return nil
}

// Wrapper around exec.ExitError that avoids the default handling by urfave/cli.
type SilentExitError struct {
	ExitCode  int
	exitError error
}

func (ee SilentExitError) Error() string {
	return ee.exitError.Error()
}
