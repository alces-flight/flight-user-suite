package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
	"github.com/yarlson/pin"
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
            	report, err := buildDoctorReport(ctx)
            	if err != nil {
            		return err
            	}
            	return writeChecksJSON(report)
            }
			greenText := lipgloss.NewStyle().Foreground(lipgloss.Green)
			redText := lipgloss.NewStyle().Foreground(lipgloss.Red)
			fmt.Println()
			allOK := checkRequiredDeps(
				ctx,
				"Checking critical dependencies...",
				"Critical dependencies",
				"Critical dependencies not satisfied",
				requiredDependencies(config.Dependencies),
			)
			checkOptionalDeps(
				ctx,
				"Checking optional dependencies...",
				"Optional dependencies",
				"OPTIONAL dependencies not satisfied",
				optionalDependencies(config.Dependencies),
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
					ok := checkRequiredDeps(
						ctx,
						fmt.Sprintf("Checking required dependencies for %s desktop type.", typ.ID),
						fmt.Sprintf("Required dependencies for %s desktop type", typ.ID),
						fmt.Sprintf("Missing required dependencies for %s desktop type", typ.ID),
						requiredDependencies(typ.dependencies),
					)
					allOK = allOK && ok
					checkOptionalDeps(
						ctx,
						fmt.Sprintf("Checking optional dependencies for %s desktop type.", typ.ID),
						fmt.Sprintf("Optional dependencies for %s desktop type", typ.ID),
						fmt.Sprintf("OPTIONAL dependencies for %s desktop type are not satisfied", typ.ID),
						optionalDependencies(typ.dependencies),
					)
				}
			}
			if allOK {
				fmt.Println()
				msg := greenText.Render("\u2705 All required dependencies satisfied")
				lipgloss.Println(msg)
				return nil
			} else {
				fmt.Println()
				msg := redText.Render("\u274c Required dependencies not satisfied")
				lipgloss.Println(msg)
				return cli.Exit("", 1)
			}
		},
	}
}

type doctorReport struct {
	OK    bool                 `json:"ok"`
	Core  depGroup             `json:"core"`
	Types []typeReport         `json:"types"`
}

type typeReport struct {
	ID       string   `json:"id"`
	Required depGroup `json:"required"`
	Optional depGroup `json:"optional"`
}

type depGroup struct {
	OK     bool              `json:"ok"`
	Checks []checkResultJSON `json:"checks"`
}

type checkResultJSON struct {
	Type           string   `json:"type"`
	Paths          []string `json:"paths"`
	Description    string   `json:"description,omitempty"`
	Optional       bool     `json:"optional"`
	Found          bool     `json:"found"`
	FoundAt        string   `json:"found_at,omitempty"`
	SuccessMessage string   `json:"success_message,omitempty"`
	FailureMessage string   `json:"failure_message,omitempty"`
	Error          string   `json:"error,omitempty"`
}

func writeChecksJSON(report doctorReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func toJSONResults(results []checkResult) []checkResultJSON {
	out := make([]checkResultJSON, 0, len(results))

	for _, r := range results {
		item := checkResultJSON{
			Type:           r.dependency.Type,
			Paths:          r.dependency.Paths,
			Description:    r.dependency.Description,
			Optional:       r.dependency.Optional,
			Found:          r.found,
			FoundAt:        r.foundAt,
			SuccessMessage: r.dependency.SuccessMessage,
			FailureMessage: r.dependency.FailureMessage,
		}

		if r.err != nil {
			item.Error = r.err.Error()
		}

		out = append(out, item)
	}

	return out
}

func buildDoctorReport(ctx context.Context) (doctorReport, error) {
	report := doctorReport{OK: true}

	// --- Core required deps
	coreRequired := requiredDependencies(config.Dependencies)
	coreResults, coreOK := runDoctor(coreRequired)

	report.Core = depGroup{
		OK:     coreOK,
		Checks: toJSONResults(coreResults),
	}

	report.OK = report.OK && coreOK

	// --- Load desktop types
	types, err := loadAllTypes(false)
	if err != nil {
		return report, err
	}

	for _, typ := range types {
		tr := typeReport{ID: typ.ID}

		if err := typ.loadDependencies(); err != nil {
			// treat as failed type
			tr.Required = depGroup{
				OK: false,
				Checks: []checkResultJSON{
					{
						Type:  typ.ID,
						Error: err.Error(),
					},
				},
			}
			report.OK = false
			report.Types = append(report.Types, tr)
			continue
		}

		// required
		reqDeps := requiredDependencies(typ.dependencies)
		reqResults, reqOK := runDoctor(reqDeps)

		tr.Required = depGroup{
			OK:     reqOK,
			Checks: toJSONResults(reqResults),
		}

		// optional
		optDeps := optionalDependencies(typ.dependencies)
		optResults, optOK := runDoctor(optDeps)

		tr.Optional = depGroup{
			OK:     optOK,
			Checks: toJSONResults(optResults),
		}

		report.OK = report.OK && reqOK
		report.Types = append(report.Types, tr)
	}

	return report, nil
}

func checkRequiredDeps(ctx context.Context, spinnerText string, doneText string, failText string, deps []dependency) bool {
	allOK := true
	p := pin.New(spinnerText,
		pin.WithSpinnerColor(pin.ColorCyan),
		pin.WithTextColor(pin.ColorGreen),
		pin.WithDoneSymbol('\u2705'),
		pin.WithFailSymbol('\u274c'),
		pin.WithFailColor(pin.ColorRed),
	)
	cancel := p.Start(ctx)
	defer cancel()
	checkResults, ok := runDoctor(deps)
	<-time.After(1 * time.Second)
	if ok {
		p.Stop(doneText)
	} else {
		p.Fail(failText)
		allOK = false
	}
	printCheckResults(checkResults)
	return allOK
}

func checkOptionalDeps(ctx context.Context, spinnerText string, doneText string, failText string, deps []dependency) {
	if len(deps) == 0 {
		return
	}
	fmt.Println()
	p := pin.New(spinnerText,
		pin.WithSpinnerColor(pin.ColorCyan),
		pin.WithTextColor(pin.ColorGreen),
		pin.WithDoneSymbol('\u2705'),
		pin.WithFailSymbol('\u274c'),
		pin.WithFailColor(pin.ColorYellow),
	)
	cancel := p.Start(ctx)
	defer cancel()
	checkResults, ok := runDoctor(deps)
	<-time.After(1 * time.Second)
	if ok {
		p.Stop(doneText)
	} else {
		p.Fail(failText)
	}
	printCheckResults(checkResults)
}

type checkResult struct {
	dependency dependency
	found      bool
	foundAt    string
	err        error
}

func runDoctor(dependencies []dependency) ([]checkResult, bool) {
	checkResults := make([]checkResult, 0)
	allOK := true
	for _, dep := range dependencies {
		switch dep.Type {
		case "exe":
			result := checkExeAvailable(dep)
			if result.err != nil {
				allOK = false
			}
			checkResults = append(checkResults, result)
		case "dir":
			result := checkDirNonEmpty(dep)
			if result.err != nil {
				allOK = false
			}
			checkResults = append(checkResults, result)
		}
	}
	return checkResults, allOK
}

func printCheckResults(checkResults []checkResult) {
	depParts := make([]string, 0, len(checkResults))
	resultParts := make([]string, 0, len(checkResults))
	for _, result := range checkResults {
		tick := " > \u2705 "
		outcome := result.foundAt
		if len(result.dependency.SuccessMessage) > 0 {
			outcome = lipgloss.Wrap(result.dependency.SuccessMessage, 60, "")
		}
		if !result.found {
			tick = " > \u274c "
			if len(result.dependency.FailureMessage) > 0 {
				outcome = lipgloss.Wrap(result.dependency.FailureMessage, 60, "")
			} else {
				outcome = result.err.Error()
			}
		}
		description := result.dependency.Description
		if description == "" {
			description = strings.Join(result.dependency.Paths, "\n")
		}
		depPart := lipgloss.JoinHorizontal(lipgloss.Top, tick, description)
		resultPart := lipgloss.JoinHorizontal(lipgloss.Top, " : ", outcome)
		depParts = append(depParts, depPart)
		resultParts = append(resultParts, resultPart)
	}

	out := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, depParts...),
		lipgloss.JoinVertical(lipgloss.Left, resultParts...),
	)
	lipgloss.Println("")
	lipgloss.Println(out)
}

func checkExeAvailable(dep dependency) checkResult {
	var errs error
	for _, path := range dep.Paths {
		location, err := exec.LookPath(path)
		if err == nil {
			return checkResult{
				dependency: dep,
				found:      true,
				foundAt:    location,
				err:        nil,
			}
		}
		errs = errors.Join(errs, err)
	}
	return checkResult{
		dependency: dep,
		found:      false,
		foundAt:    "",
		err:        errs,
	}
}

func checkDirNonEmpty(dep dependency) checkResult {
	var errs error
	nonEmptyDirs := make([]string, 0)
	for _, dir := range dep.Paths {
		entries, err := os.ReadDir(dir)
		if len(entries) > 0 {
			nonEmptyDirs = append(nonEmptyDirs, dir)
		} else {
			errs = errors.Join(errs, err)
		}
	}
	if len(nonEmptyDirs) > 0 {
		return checkResult{
			dependency: dep,
			found:      true,
			foundAt:    strings.Join(nonEmptyDirs, "\n"),
			err:        nil,
		}
	}
	return checkResult{
		dependency: dep,
		found:      false,
		foundAt:    "",
		err:        errs,
	}
}
