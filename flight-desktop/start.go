package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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

func printSessionDetails(session Session) {
	// TODO: Better output for TTY.
	fmt.Println()
	fmt.Printf("Identity\t%s\n", session.UUID)
	fmt.Printf("Name\t%s\n", session.Name)
	fmt.Printf("Type\t%s\n", session.SessionType)
	fmt.Printf("Password\t%s\n", session.Password)
	fmt.Printf("State\t%s\n", session.SessionState)
	fmt.Printf("Created at\t%s\n", session.CreatedAt)
	fmt.Printf("Geometry\t%s\n", session.Geometry)
	fmt.Println()
}

type sessionState string

var (
	New    sessionState = "new"
	Active sessionState = "active"
)

type Session struct {
	UUID         uuid.UUID `yaml:"uuid"`
	Name         string    `yaml:"name"`
	SessionType  string    `yaml:"session_type"`
	Password     string
	SessionState sessionState `yaml:"session_state"`
	Geometry     string       `yaml:"geometry"`
	CreatedAt    time.Time    `yaml:"created_at"`
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
			log.Debug("vncpasswd output", "stdout", output, "stderr", string(ee.Stderr))
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
	// TODO: Use correct script
	vncServerScriptPath := filepath.Join(flightRoot, "usr", "libexec", "vncserver")
	passwdFile := s.passwordFile()
	args := []string{
		"-autokill",
		"-sessiondir", s.sessionDir(),
		"-sessionscript", s.sessionScript(),
		"-vncpasswd", passwdFile,
		"-exedir", "/usr/bin",
		"-geometry", s.Geometry,
	}
	cmd := exec.CommandContext(ctx, vncServerScriptPath, args...)
	cmd.Dir = dir
	// TODO: Set environment.
	// cmd.Env = []string{}

	output, err := cmd.Output()
	if err != nil {
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
			log.Debug("vncserver", "stdout", output, "stderr", string(ee.Stderr))
		}
		return err
	}
	fmt.Printf("Output >>>\n%s\n", string(output))
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
