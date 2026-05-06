package main

import (
	"fmt"
	"net/http"

	"github.com/concertim/flight-user-suite/flight-web-suite/desktop"
	"github.com/concertim/flight-user-suite/flight/toolset"
	"github.com/labstack/echo/v5"
)

type desktopGeometryOption struct {
	Value   string
	Label   string
	Default bool
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
	Alert       string
}

var desktopGeometryOptions = []desktopGeometryOption{
	{Value: "1280x720", Label: "1280 x 720"},
	{Value: "1024x768", Label: "1024 x 768 (default)", Default: true},
	{Value: "1600x900", Label: "1600 x 900"},
	{Value: "1280x1024", Label: "1280 x 1024"},
	{Value: "1920x1080", Label: "1920 x 1080"},
	{Value: "1600x1200", Label: "1600 x 1200"},
}

func newDesktopSessionHandler(c *echo.Context) error {
	if !IsLoggedIn(c) {
		return c.Redirect(http.StatusSeeOther, "/sessions")
	}
	if err := requireDesktopToolEnabled(); err != nil {
		return err
	}

	dependencyReport, err := desktop.DoctorCommand(c.Request().Context(), env, CurrentUserName(c))
	if err != nil {
		return err
	}

	desktopTypes, err := desktop.AvailCommand(c.Request().Context(), env, CurrentUserName(c))
	if err != nil {
		return err
	}

	form := desktopLaunchFormData{
		DesktopType: defaultDesktopType(desktopTypes),
		Geometry:    defaultDesktopGeometry(),
	}
	return renderDesktopLaunchPage(c, http.StatusOK, desktopTypes, dependencyReport, form)
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

	desktopTypes, err := desktop.AvailCommand(c.Request().Context(), env, CurrentUserName(c))
	if err != nil {
		return err
	}

	validateDesktopLaunchForm(desktopTypes, &form)
	if form.Errors == (desktopLaunchFieldErrors{}) {
		response, err := desktop.StartCommand(c.Request().Context(), env, CurrentUserName(c), desktop.StartInput{
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
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/desktop/%s", response.SessionName))
		}
		applyDesktopStartErrors(&form, response)
	}

	sess, err := GetSession(c)
	if err != nil {
		return err
	}
	alert := form.Alert
	if alert == "" {
		alert = "Failed to launch desktop session."
	}
	sess.AddFlash(alert, "alert")
	SaveSession(c, sess)
	return renderDesktopLaunchPage(c, http.StatusUnprocessableEntity, desktopTypes, nil, form)
}

func requireDesktopToolEnabled() error {
	tool, err := toolset.GetTool(env.FlightRoot, "desktop")
	if err != nil || !tool.Enabled {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Flight Desktop is not enabled")
	}
	return nil
}

func validateDesktopLaunchForm(desktopTypes []*desktop.Type, form *desktopLaunchFormData) {
	if !hasDesktopType(desktopTypes, form.DesktopType) {
		form.Errors.DesktopType = "Select an available desktop type."
	}
	if !hasDesktopGeometry(form.Geometry) {
		form.Errors.Geometry = "Select a supported geometry."
	}
}

func applyDesktopStartErrors(form *desktopLaunchFormData, response desktop.StartResponse) {
	form.Alert = "Failed to launch desktop session."
	if response.Error.Code == "dependencies_failed" {
		form.Alert = response.Error.Detail
	}
	switch {
	case response.Error.Source != nil && response.Error.Source.Parameter == "type":
		form.Errors.DesktopType = response.Error.Detail
	case response.Error.Source != nil && response.Error.Source.Parameter == "name":
		form.Errors.Name = response.Error.Detail
	case response.Error.Source != nil && response.Error.Source.Parameter == "geometry":
		form.Errors.Geometry = response.Error.Detail
	}
}

func renderDesktopLaunchPage(c *echo.Context, status int, desktopTypes []*desktop.Type, doctorReport *desktop.DoctorReport, form desktopLaunchFormData) error {
	if form.DesktopType == "" {
		form.DesktopType = defaultDesktopType(desktopTypes)
	}
	if form.Geometry == "" {
		form.Geometry = defaultDesktopGeometry()
	}
	coreDependenciesOK := false
	missingDependencies := []string{}
	if doctorReport != nil {
	    coreDependenciesOK = doctorReport.OK
	    for _, dep := range doctorReport.Checks {
	        if dep.Found != true {
	            missingDependencies = append(missingDependencies, dep.Description)
	        }
	    }
	}

	data := map[string]any{
	    "CoreDependenciesOK":       coreDependenciesOK,
	    "MissingDependencies":      missingDependencies,
		"DesktopTypes":             desktopTypes,
		"NumAvailableDesktopTypes": numAvailableDesktopTypes(desktopTypes),
		"GeometryOptions":          desktopGeometryOptions,
		"Form":                     form,
	}
	return c.Render(status, "desktop/new", data)
}

func defaultDesktopType(desktopTypes []*desktop.Type) string {
	if len(desktopTypes) == 0 {
		return ""
	}
	for _, typ := range desktopTypes {
		if typ.IsAvailable {
			return typ.ID
		}
	}
	return ""
}

func numAvailableDesktopTypes(desktopTypes []*desktop.Type) int {
	numAvailable := 0
	for _, typ := range desktopTypes {
		if typ.IsAvailable {
			numAvailable++
		}
	}
	return numAvailable
}

func defaultDesktopGeometry() string {
	for _, option := range desktopGeometryOptions {
		if option.Default {
			return option.Value
		}
	}
	return ""
}

func hasDesktopType(desktopTypes []*desktop.Type, id string) bool {
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
