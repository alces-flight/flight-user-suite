package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/concertim/flight-user-suite/flight/toolset"
	"github.com/labstack/echo/v5"
)

type desktopType struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	URL     string `json:"url"`
}

type desktopGeometryOption struct {
	Value   string
	Label   string
	Default bool
}

type desktopStartInput struct {
	DesktopType string
	Name        string
	Geometry    string
}

type desktopStartResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error"`
	Reason      string `json:"reason"`
}

type desktopLaunchFieldErrors struct {
	DesktopType string
	Name        string
	Geometry    string
}

type desktopLaunchFormData struct {
	DesktopType string
	Name        string
	Geometry    string
	Errors      desktopLaunchFieldErrors
	FormError   string
}

var desktopGeometryOptions = []desktopGeometryOption{
	{Value: "1024x768", Label: "1024 x 768 (default)", Default: true},
	{Value: "1280x1024", Label: "1280 x 1024"},
}

func newDesktopSessionHandler(c *echo.Context) error {
	if !IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/sessions")
	}
	if err := requireDesktopToolEnabled(); err != nil {
		return err
	}

	desktopTypes, err := desktopAvailCommand(c.Request().Context(), CurrentUserName(c))
	if err != nil {
		return err
	}

	form := desktopLaunchFormData{
		DesktopType: defaultDesktopType(desktopTypes),
		Geometry:    defaultDesktopGeometry(),
	}
	return renderDesktopLaunchPage(c, http.StatusOK, desktopTypes, form)
}

func createDesktopSessionHandler(c *echo.Context) error {
	if !IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/sessions")
	}
	if err := requireDesktopToolEnabled(); err != nil {
		return err
	}

	form := desktopLaunchFormData{
		DesktopType: c.FormValue("desktop_type"),
		Name:        c.FormValue("name"),
		Geometry:    c.FormValue("geometry"),
	}

	desktopTypes, err := desktopAvailCommand(c.Request().Context(), CurrentUserName(c))
	if err != nil {
		return err
	}

	validateDesktopLaunchForm(desktopTypes, &form)
	if form.Errors == (desktopLaunchFieldErrors{}) {
		response, err := desktopStartCommand(c.Request().Context(), CurrentUserName(c), desktopStartInput{
			DesktopType: form.DesktopType,
			Name:        form.Name,
			Geometry:    form.Geometry,
		})
		if err != nil {
			return err
		}
		if response.Success {
			sess, err := GetSession(c)
			if err != nil {
				return err
			}
			sess.AddFlash(fmt.Sprintf("Desktop session '%s' launched.", response.SessionName), "notice")
			SaveSession(c, sess)
			return c.Redirect(http.StatusSeeOther, "/desktop")
		}
		applyDesktopStartErrors(&form, response)
	}

	sess, err := GetSession(c)
	if err != nil {
		return err
	}
	alertMessage := form.FormError
	if alertMessage == "" {
		alertMessage = "Failed to launch desktop session."
	}
	sess.AddFlash(alertMessage, "alert")
	SaveSession(c, sess)
	return renderDesktopLaunchPage(c, http.StatusUnprocessableEntity, desktopTypes, form)
}

func requireDesktopToolEnabled() error {
	tool, err := toolset.GetTool(env.FlightRoot, "desktop")
	if err != nil || !tool.Enabled {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Flight Desktop is not enabled")
	}
	return nil
}

func desktopAvailCommand(ctx context.Context, username string) ([]desktopType, error) {
	cmd, err := buildDesktopCommand(ctx, username, "avail", "--format", "json")
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

	var desktopTypes []desktopType
	if err := json.Unmarshal(stdout.Bytes(), &desktopTypes); err != nil {
		return nil, fmt.Errorf("decoding desktop types: %w", err)
	}
	slices.SortFunc(desktopTypes, func(a, b desktopType) int {
		return strings.Compare(a.ID, b.ID)
	})
	return desktopTypes, nil
}

func desktopStartCommand(ctx context.Context, username string, input desktopStartInput) (desktopStartResponse, error) {
	args := []string{"start", input.DesktopType, "--format", "json", "--geometry", input.Geometry}
	if input.Name != "" {
		args = append(args, "--name", input.Name)
	}

	cmd, err := buildDesktopCommand(ctx, username, args...)
	if err != nil {
		return desktopStartResponse{}, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	var response desktopStartResponse
	if decodeErr := json.Unmarshal(stdout.Bytes(), &response); decodeErr == nil {
		return response, nil
	}
	if runErr != nil {
		if stderr.Len() != 0 {
			return desktopStartResponse{}, fmt.Errorf("starting desktop session: %s", stderr.String())
		}
		return desktopStartResponse{}, fmt.Errorf("starting desktop session: %w", runErr)
	}
	return desktopStartResponse{}, fmt.Errorf("decoding desktop start response: %s", stdout.String())
}

func validateDesktopLaunchForm(desktopTypes []desktopType, form *desktopLaunchFormData) {
	if !hasDesktopType(desktopTypes, form.DesktopType) {
		form.Errors.DesktopType = "Select an available desktop type."
	}
	if !hasDesktopGeometry(form.Geometry) {
		form.Errors.Geometry = "Select a supported geometry."
	}
}

func applyDesktopStartErrors(form *desktopLaunchFormData, response desktopStartResponse) {
	switch response.Reason {
	case "invalid_type":
		form.Errors.DesktopType = response.Error
	case "invalid_name":
		form.Errors.Name = response.Error
	case "invalid_geometry":
		form.Errors.Geometry = response.Error
	default:
		form.FormError = response.Error
	}
}

func renderDesktopLaunchPage(c *echo.Context, status int, desktopTypes []desktopType, form desktopLaunchFormData) error {
	if form.DesktopType == "" {
		form.DesktopType = defaultDesktopType(desktopTypes)
	}
	if form.Geometry == "" {
		form.Geometry = defaultDesktopGeometry()
	}

	data := map[string]any{
		"DesktopTypes":    desktopTypes,
		"GeometryOptions": desktopGeometryOptions,
		"Form":            form,
	}
	return c.Render(status, "desktop/new", AddCommonData(c, data))
}

func defaultDesktopType(desktopTypes []desktopType) string {
	if len(desktopTypes) == 0 {
		return ""
	}
	return desktopTypes[0].ID
}

func defaultDesktopGeometry() string {
	for _, option := range desktopGeometryOptions {
		if option.Default {
			return option.Value
		}
	}
	return ""
}

func hasDesktopType(desktopTypes []desktopType, id string) bool {
	for _, desktopType := range desktopTypes {
		if desktopType.ID == id {
			return true
		}
	}
	return false
}

func hasDesktopGeometry(geometry string) bool {
	for _, option := range desktopGeometryOptions {
		if option.Value == geometry {
			return true
		}
	}
	return false
}
