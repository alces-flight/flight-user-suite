package desktop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/concertim/flight-user-suite/flight/configenv"
)

type Type struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	URL         string `json:"url"`
	IsAvailable bool   `json:"available"`
}

type StartInput struct {
	DesktopType string
	Name        string
	Geometry    string
}

type StartResponse struct {
	Success     bool
	SessionName string
	Error       startError
}

type startError struct {
	Code   string            `json:"code"`
	Title  string            `json:"title"`
	Detail string            `json:"detail"`
	Source *startErrorSource `json:"source,omitempty"`
}

type startErrorSource struct {
	Parameter string `json:"parameter"`
}

type startSuccessDocument struct {
	Success     *bool  `json:"success"`
	SessionName string `json:"session_name"`
}

type startErrorDocument struct {
	Errors []startError `json:"errors"`
}

func AvailCommand(ctx context.Context, env configenv.Env, username string) ([]*Type, error) {
	cmd, err := buildLocalDesktopCommand(ctx, env, username, "avail", "--format", "json")
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() != 0 {
			return nil, fmt.Errorf("listing desktop types: %s", stderr.String())
		}
		return nil, fmt.Errorf("listing desktop types: %w", err)
	}
	var desktopTypes []*Type
	if err := json.Unmarshal(stdout.Bytes(), &desktopTypes); err != nil {
		return nil, fmt.Errorf("decoding desktop types: %w", err)
	}
	slices.SortFunc(desktopTypes, func(a, b *Type) int {
		return strings.Compare(a.ID, b.ID)
	})
	return desktopTypes, nil
}

type DoctorReport struct {
	OK     bool              `json:"ok"`
	Checks []CheckResultJSON `json:"checks"`
}

type CheckResultJSON struct {
	Description string   `json:"description"`
	Paths       []string `json:"paths"`
	Optional    bool     `json:"optional"`
	Found       bool     `json:"found"`
	Error       string   `json:"error"`
}

func DoctorCommand(ctx context.Context, env configenv.Env, username string) (*DoctorReport, error) {
	cmd, err := buildLocalDesktopCommand(ctx, env, username, "doctor", "--format", "json")
	if err != nil {
		return nil, err
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() != 0 {
			return nil, fmt.Errorf("running desktop doctor: %s", stderr.String())
		}
		return nil, fmt.Errorf("running desktop doctor: %w", err)
	}
	var dependencyReport DoctorReport
	if err := json.Unmarshal(stdout.Bytes(), &dependencyReport); err != nil {
		return nil, fmt.Errorf("decoding dependency checks: %w", err)
	}
	return &dependencyReport, nil
}

func StartCommand(ctx context.Context, logger *slog.Logger, env configenv.Env, username string, input StartInput) (StartResponse, error) {
	// TODO:
	// * Add support for round robin over remote hosts.
	// E.g., FWS is installed on infra node. Sessions it starts should be installed on `login1`, `login2`, etc..
	// * If remote launch is configured into remote host and launch it. Otherwise, run command locally.
	// * Ensure we use the same host as DoctorCommand. Perhaps a hidden form field.
	args := []string{"start", input.DesktopType, "--format", "json", "--geometry", input.Geometry}
	if input.Name != "" {
		args = append(args, "--name", input.Name)
	}

	logger.Info("DESKTOP SESSION", "action", "start", "name", input.Name, "username", username, "type", input.DesktopType, "remote", false)
	cmd, err := buildLocalDesktopCommand(ctx, env, username, args...)
	if err != nil {
		return StartResponse{}, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	var successDoc startSuccessDocument
	if decodeErr := json.Unmarshal(stdout.Bytes(), &successDoc); decodeErr == nil && successDoc.Success != nil {
		return StartResponse{
			Success:     *successDoc.Success,
			SessionName: successDoc.SessionName,
		}, nil
	}

	var errorDoc startErrorDocument
	if decodeErr := json.Unmarshal(stdout.Bytes(), &errorDoc); decodeErr == nil && len(errorDoc.Errors) > 0 {
		return StartResponse{
			Success: false,
			Error:   errorDoc.Errors[0],
		}, nil
	}
	if runErr != nil {
		if stderr.Len() != 0 {
			return StartResponse{}, fmt.Errorf("starting desktop session: %s", stderr.String())
		}
		return StartResponse{}, fmt.Errorf("starting desktop session: %w", runErr)
	}
	return StartResponse{}, fmt.Errorf("decoding desktop start response: %s", stdout.String())
}
