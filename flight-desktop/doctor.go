package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/concertim/flight-user-suite/flight/doctor"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func doctorCommand() *cli.Command {
	return &cli.Command{
		Name:        "doctor",
		Usage:       "System health check",
		Description: wordwrap.String("Perform a health check on the system to check that all dependencies are present.", maxTextWidth),
		Category:    "Desktop types",
		Flags:       []cli.Flag{formatFlag},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.String("format") == "json" {
				report := buildDoctorReport()
				return writeChecksJSON(report)
			}
			greenText := lipgloss.NewStyle().Foreground(lipgloss.Green)
			redText := lipgloss.NewStyle().Foreground(lipgloss.Red)
			fmt.Println()
			allOK := doctor.CheckRequiredDeps(
				ctx,
				"Checking critical dependencies...",
				"Critical dependencies",
				"Critical dependencies not satisfied",
				doctor.RequiredDependencies(config.Dependencies),
			)
			doctor.CheckOptionalDeps(
				ctx,
				"Checking optional dependencies...",
				"Optional dependencies",
				"OPTIONAL dependencies not satisfied",
				doctor.OptionalDependencies(config.Dependencies),
			)

			types, err := loadAllTypes(false)
			if err != nil {
				fmt.Print("\u274c ")
				lipgloss.Println(redText.Render("Checking dependencies for desktop types failed"))
				fmt.Printf("\n > %s\n", err)
				allOK = false
			} else {
				for _, typ := range types {
					fmt.Println()
					if err := typ.loadDependencies(); err != nil {
						fmt.Print("\u274c ")
						lipgloss.Println(redText.Render(fmt.Sprintf("Checking dependencies for %s desktop type failed", typ.ID)))
						fmt.Printf("\n > %s\n", err)
						allOK = false
						continue
					}
					ok := doctor.CheckRequiredDeps(
						ctx,
						fmt.Sprintf("Checking required dependencies for %s desktop type.", typ.ID),
						fmt.Sprintf("Required dependencies for %s desktop type", typ.ID),
						fmt.Sprintf("Missing required dependencies for %s desktop type", typ.ID),
						doctor.RequiredDependencies(typ.dependencies),
					)
					allOK = allOK && ok
					doctor.CheckOptionalDeps(
						ctx,
						fmt.Sprintf("Checking optional dependencies for %s desktop type.", typ.ID),
						fmt.Sprintf("Optional dependencies for %s desktop type", typ.ID),
						fmt.Sprintf("OPTIONAL dependencies for %s desktop type are not satisfied", typ.ID),
						doctor.OptionalDependencies(typ.dependencies),
					)
				}
			}
			if allOK {
				fmt.Println()
				msg := greenText.Render("\u2705 All required dependencies satisfied")
				lipgloss.Println(msg)
				return nil
			}

			fmt.Println()
			msg := redText.Render("\u274c Required dependencies not satisfied")
			lipgloss.Println(msg)
			return cli.Exit("", 1)
		},
	}
}

type doctorReport struct {
	OK     bool              `json:"ok"`
	Checks []checkResultJSON `json:"checks"`
}

type checkResultJSON struct {
	Description string   `json:"description,omitempty"`
	Paths       []string `json:"paths"`
	Optional    bool     `json:"optional"`
	Found       bool     `json:"found"`
	Error       string   `json:"error,omitempty"`
}

func writeChecksJSON(report doctorReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func toJSONResults(results []doctor.CheckResult) []checkResultJSON {
	jsonResults := make([]checkResultJSON, 0, len(results))
	for _, r := range results {
		item := checkResultJSON{
			Description: r.Dependency.Description,
			Paths:       r.Dependency.Paths,
			Optional:    r.Dependency.Optional,
			Found:       r.Found,
		}
		if r.Err != nil {
			item.Error = r.Err.Error()
		}
		jsonResults = append(jsonResults, item)
	}
	return jsonResults
}

func buildDoctorReport() doctorReport {
	coreRequired := doctor.RequiredDependencies(config.Dependencies)
	coreResults, coreOK := doctor.Run(coreRequired)
	return doctorReport{
		OK:     coreOK,
		Checks: toJSONResults(coreResults),
	}
}
