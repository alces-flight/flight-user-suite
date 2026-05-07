package main

import (
	"context"
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/concertim/flight-user-suite/flight/doctor"
	"github.com/concertim/flight-user-suite/flight/userroles"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

var webDoctorGeneralDependencies = []doctor.Dependency{
	{
		Type:        doctor.TypeExecutable,
		Description: "Python 3",
		Paths:       []string{"python3"},
	},
	{
		Type:           doctor.TypePythonModule,
		Description:    "Python PAM",
		Paths:          []string{"python3"},
		Module:         "pam",
		FailureMessage: "Python PAM module (pam) not found",
	},
}

func doctorCommand() *cli.Command {
	return &cli.Command{
		Name:        "doctor",
		Usage:       "System health check for Flight Web Suite dependencies",
		Description: wordwrap.String("Perform a health check on the system to check that all Flight Web Suite dependencies are present.", maxTextWidth),
		Action:      userroles.AsRoot(runWebDoctor),
	}
}

func runWebDoctor(ctx context.Context, cmd *cli.Command) error {
	greenText := lipgloss.NewStyle().Foreground(lipgloss.Green)
	redText := lipgloss.NewStyle().Foreground(lipgloss.Red)

	fmt.Println()
	allOK := doctor.CheckRequiredDeps(
		ctx,
		"Checking required general Flight Web Suite dependencies...",
		"Required general Flight Web Suite dependencies",
		"Required general Flight Web Suite dependencies not satisfied",
		doctor.RequiredDependencies(webDoctorGeneralDependencies),
	)

	fmt.Println()
	if allOK {
		lipgloss.Println(greenText.Render("\u2705 All required dependencies satisfied"))
		return nil
	}

	lipgloss.Println(redText.Render("\u274c Required dependencies not satisfied"))
	return cli.Exit("", 1)
}
