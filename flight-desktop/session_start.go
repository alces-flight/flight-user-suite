package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
	"github.com/yarlson/pin"
)

var (
	nameWhitelist            = "-_.A-Za-z0-9"
	nameWhitelistExplanation = "letters, numbers, hyphens, underscores and dots"
	nameBlacklist            = regexp.MustCompile(fmt.Sprintf("^-|[^%s]+", nameWhitelist))
	nameMaxLen               = 40
)

func libexecPath(relpath string) string {
	return filepath.Join(env.FlightRoot, "usr", "libexec", "desktop", relpath)
}

func startSessionCommand() *cli.Command {
	return &cli.Command{
		Name:        "start",
		Usage:       "Start an interactive desktop session",
		Description: wordwrap.String("Start a new interactive desktop session and display details about the new session.\n\nAvailable desktop types can be shown using the 'avail' command.", maxTextWidth),
		Category:    "Sessions",
		Flags: []cli.Flag{
			formatFlag,
			&cli.StringFlag{
				Name:        "name",
				Aliases:     []string{"n"},
				Usage:       "Name the desktop session `NAME` so it can be more easily identified.",
				DefaultText: "random",
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
		Before: assertArgPresent("type"),
		ShellComplete: func(ctx context.Context, cmd *cli.Command) {
			cli.DefaultCompleteWithFlags(ctx, cmd)
			switch cmd.NArg() {
			case 0:
				types, err := loadAllTypes(false)
				if err != nil {
					return
				}
				for _, t := range types {
					fmt.Println(t.ID)
				}
			}
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			sessionType := cmd.StringArg("type")
			nameInput := cmd.String("name")
			geometry := cmd.String("geometry")
			if cmd.String("format") == "json" {
				return startSessionJSON(ctx, sessionType, nameInput, geometry)
			}
			return startSessionPretty(ctx, sessionType, nameInput, geometry)
		},
	}
}

func startSessionPretty(ctx context.Context, sessionType, nameInput, geometry string) error {
	if err := validateSessionType(sessionType); err != nil {
		return err
	}
	if err := validateSessionName(nameInput); err != nil {
		return err
	}
	if err := validateGeometry(geometry); err != nil {
		return err
	}
	name := nameInput
	if name == "" {
		name = newNameGenerator(sessionType).Generate()
	}
	fmt.Printf("Starting '%s' desktop session '%s':\n\n", sessionType, name)

	depsOK, err := checkDependencies(ctx, sessionType)

	if !depsOK {
		return err
	}

	p := createPin("Starting session...")
	cancel := p.Start(ctx)
	defer cancel()

	session := Session{
		Name:        name,
		State:       New,
		SessionType: sessionType,
		Geometry:    geometry,
	}
	err = session.Start(ctx)
	if err != nil {
		p.Fail("Starting session failed")
		return err
	}
	p.Stop(fmt.Sprintf("Your %s session is ready!", session.SessionType))
	fmt.Println()
	sessionStarted(&session)
	connectionInfo(&session)
	managementInfo(&session)
	return nil
}

type sessionStartResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
}

type sessionStartErrorDocument struct {
	Errors []sessionStartError `json:"errors"`
}

type sessionStartError struct {
	Code   string                   `json:"code"`
	Title  string                   `json:"title"`
	Detail string                   `json:"detail"`
	Source *sessionStartErrorSource `json:"source,omitempty"`
	Meta   sessionStartErrorMeta    `json:"meta"`
}

type sessionStartErrorSource struct {
	Parameter string `json:"parameter"`
}

type sessionStartErrorMeta struct {
	CLIDetail string `json:"cli_detail"`
}

type startValidationError struct {
	message string // Human-readable error message.
	reason  string // Machine-readable error message.
	title   string
	detail  string
	source  string
}

func (e startValidationError) Error() string {
	return e.message
}

func startSessionJSON(ctx context.Context, sessionType, nameInput, geometry string) error {
	if err := validateSessionType(sessionType); err != nil {
		return writeStartFailure(startFailureError(err, nameInput))
	}
	if err := validateSessionName(nameInput); err != nil {
		return writeStartFailure(startFailureError(err, nameInput))
	}
	if err := validateGeometry(geometry); err != nil {
		return writeStartFailure(startFailureError(err, nameInput))
	}

	sessionName := nameInput
	if sessionName == "" {
		sessionName = newNameGenerator(sessionType).Generate()
	}

	depsOK, err := checkDependenciesQuiet(sessionType)
	if err != nil {
		return writeStartFailure(startFailureError(err, nameInput))
	}
	if !depsOK {
		return writeStartFailure(sessionStartError{
			Code:   "dependencies_failed",
			Title:  "Desktop dependencies unavailable",
			Detail: fmt.Sprintf("Missing required dependencies for %s desktop type.", sessionType),
			Meta: sessionStartErrorMeta{
				CLIDetail: fmt.Sprintf("Missing required dependencies for %s desktop type.", sessionType),
			},
		})
	}

	session := Session{
		Name:        sessionName,
		State:       New,
		SessionType: sessionType,
		Geometry:    geometry,
	}
	if err := session.Start(ctx); err != nil {
		return writeStartFailure(startFailureError(err, nameInput))
	}
	return writeStartSuccess(sessionName)
}

func validateSessionType(sessionType string) error {
	availableTypes, err := loadAllTypes(false)
	if err != nil {
		return err
	}
	typeNames := make([]string, 0, len(availableTypes))
	for _, typ := range availableTypes {
		typeNames = append(typeNames, typ.ID)
	}
	if !slices.Contains(typeNames, sessionType) {
		return startValidationError{
			message: fmt.Sprintf(
				"unknown type '%s'. Valid values are %s.",
				sessionType,
				strings.Join(typeNames, ", "),
			),
			reason: "invalid_type",
			title:  "Invalid desktop type",
			detail: fmt.Sprintf(
				"Unknown desktop type '%s'. Valid values are %s.",
				sessionType,
				strings.Join(typeNames, ", "),
			),
			source: "type",
		}
	}
	return nil
}

func validateSessionName(name string) error {
	if name == "" {
		return nil
	}
	if nameBlacklist.MatchString(name) {
		return startValidationError{
			message: fmt.Sprintf("invalid value %q for flag -name: it can contain only %s and cannot start with a hyphen.", name, nameWhitelistExplanation),
			reason:  "invalid_name",
			title:   "Invalid session name",
			detail:  fmt.Sprintf("Session name can contain only %s and cannot start with a hyphen.", nameWhitelistExplanation),
			source:  "name",
		}
	}
	if len(name) > nameMaxLen {
		return startValidationError{
			message: fmt.Sprintf("invalid value %q for flag -name: it must be no more than %d characters", name, nameMaxLen),
			reason:  "invalid_name",
			title:   "Invalid session name",
			detail:  fmt.Sprintf("Session name must be no more than %d characters.", nameMaxLen),
			source:  "name",
		}
	}
	return nil
}

var geometryPattern = regexp.MustCompile(`^[1-9]\d*x[1-9]\d*$`)

func validateGeometry(geometry string) error {
	if geometryPattern.MatchString(geometry) {
		return nil
	}
	return startValidationError{
		message: fmt.Sprintf("invalid value %q for flag -geometry: it must be in WIDTHxHEIGHT format with positive integers", geometry),
		reason:  "invalid_geometry",
		title:   "Invalid desktop geometry",
		detail:  "Desktop geometry must be in WIDTHxHEIGHT format with positive integers.",
		source:  "geometry",
	}
}

func checkDependencies(ctx context.Context, sessionType string) (bool, error) {
	p := createPin("Checking system dependencies...")
	cancel := p.Start(ctx)
	defer cancel()

	// Add a small delay to stop the spinner from flickering
	<-time.After(1 * time.Second)

	globalResults, globalDepsOK := runDoctor(requiredDependencies(config.Dependencies))

	if !globalDepsOK {
		p.Fail("Missing critical dependencies")
		printCheckResults(globalResults)
		return false, nil
	}

	sessionTypeDef, err := loadType(sessionType, false)

	if err != nil {
		return false, err
	}

	err = sessionTypeDef.loadDependencies()
	if err != nil {
		return false, err
	}

	typeResults, typeDepsOK := runDoctor(requiredDependencies(sessionTypeDef.dependencies))

	if !typeDepsOK {
		p.Fail(fmt.Sprintf("Missing required dependencies for %s desktop type", sessionType))
		printCheckResults(typeResults)
		return false, err
	}

	p.Stop("Dependencies OK")

	return true, err
}

func checkDependenciesQuiet(sessionType string) (bool, error) {
	_, globalDepsOK := runDoctor(requiredDependencies(config.Dependencies))
	if !globalDepsOK {
		return false, nil
	}

	sessionTypeDef, err := loadType(sessionType, false)
	if err != nil {
		return false, err
	}
	if err := sessionTypeDef.loadDependencies(); err != nil {
		return false, err
	}

	_, typeDepsOK := runDoctor(requiredDependencies(sessionTypeDef.dependencies))
	if !typeDepsOK {
		return false, nil
	}
	return true, nil
}

func startFailureMessage(nameInput string, err error) string {
	var validationErr startValidationError
	switch {
	case errors.As(err, &validationErr):
		return err.Error()
	case strings.Contains(err.Error(), "Session name"):
		return err.Error()
	case strings.Contains(err.Error(), "geometry") && strings.Contains(err.Error(), "invalid"):
		return err.Error()
	case nameInput == "":
		return "Starting desktop session failed."
	default:
		return fmt.Sprintf("Starting desktop session '%s' failed.", nameInput)
	}
}

func startFailureReason(err error, fallback string) string {
	var validationErr startValidationError
	switch {
	case errors.As(err, &validationErr):
		return validationErr.reason
	case strings.Contains(err.Error(), "Session name"):
		return "invalid_name"
	case strings.Contains(err.Error(), "geometry") && strings.Contains(err.Error(), "invalid"):
		return "invalid_geometry"
	default:
		return fallback
	}
}

func startFailureError(err error, nameInput string) sessionStartError {
	if validationErr, ok := errors.AsType[startValidationError](err); ok {
		startErr := sessionStartError{
			Code:   validationErr.reason,
			Title:  validationErr.title,
			Detail: validationErr.detail,
			Meta: sessionStartErrorMeta{
				CLIDetail: validationErr.message,
			},
		}
		if validationErr.source != "" {
			startErr.Source = &sessionStartErrorSource{Parameter: validationErr.source}
		}
		return startErr
	}

	fallbackReason := "start_failed"
	reason := startFailureReason(err, fallbackReason)
	detail := startFailureMessage(nameInput, err)
	title := "Desktop session failed to start"
	if reason == "dependencies_failed" {
		title = "Desktop dependencies unavailable"
	}
	return sessionStartError{
		Code:   reason,
		Title:  title,
		Detail: detail,
		Meta: sessionStartErrorMeta{
			CLIDetail: err.Error(),
		},
	}
}

func writeStartSuccess(sessionName string) error {
	return writeStartResponse(sessionStartResponse{
		Success:     true,
		SessionName: sessionName,
	}, 0)
}

func writeStartFailure(startErr sessionStartError) error {
	return writeStartResponse(sessionStartErrorDocument{
		Errors: []sessionStartError{startErr},
	}, 1)
}

func writeStartResponse(response any, exitCode int) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(response); err != nil {
		return err
	}
	if exitCode == 0 {
		return nil
	}
	return SilentExitError{
		ExitCode:  exitCode,
		exitError: errors.New("session start failed"),
	}
}

func createPin(text string) *pin.Pin {
	return pin.New(
		text,
		pin.WithSpinnerColor(pin.ColorCyan),
		pin.WithTextColor(pin.ColorGreen),
		pin.WithDoneSymbol('\u2705'),
		pin.WithFailSymbol('\u274c'),
		pin.WithFailColor(pin.ColorRed),
	)
}
