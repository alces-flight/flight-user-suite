package main

import (
	"context"
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
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := loadConfig()
			if err != nil {
				return err
			}
			allOK := true
			fmt.Println()
			p := pin.New("Checking core dependencies...",
				pin.WithSpinnerColor(pin.ColorCyan),
				pin.WithTextColor(pin.ColorGreen),
				pin.WithDoneSymbol('\u2705'),
				pin.WithFailSymbol('\u274c'),
				pin.WithFailColor(pin.ColorRed),
			)
			cancel := p.Start(ctx)
			defer cancel()
			checkResults, ok := runDoctor(config.Dependencies)
			<-time.After(1 * time.Second)
			if ok {
				p.Stop("Core dependencies")
			} else {
				p.Fail("Missing core dependencies")
				allOK = false
			}
			printCheckResults(checkResults)

			types, err := loadAllTypes()
			if err != nil {
				fmt.Printf("\u274c Checking dependencies for desktop types failed: %s", err.Error())
				allOK = false
			} else {
				for _, typ := range types {
					deps, err := typ.Dependencies()
					if err != nil {
						fmt.Printf("\u274c Checking dependencies for %s desktop type failed: %s", typ.ID, err.Error())
						allOK = false
						continue
					}
					fmt.Println()
					p := pin.New(fmt.Sprintf("Checking dependencies for %s.. desktop type.", typ.ID),
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
						p.Stop(fmt.Sprintf("Dependencies for %s desktop type", typ.ID))
					} else {
						p.Fail(fmt.Sprintf("Missing dependencies for %s desktop type", typ.ID))
						allOK = false
					}
					printCheckResults(checkResults)
				}
			}
			if allOK {
				return nil
			} else {
				fmt.Println()
				return cli.Exit("Required dependencies missing", 1)
			}
		},
	}
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
		if !result.found {
			tick = " > \u274c "
			outcome = result.err.Error()
		}
		formattedPath := strings.Join(result.dependency.Paths, "\n")
		depPart := lipgloss.JoinHorizontal(lipgloss.Top, tick, formattedPath)
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
	foundNonEmtpy := false
	var foundAt string

	for _, dir := range dep.Paths {
		entries, _ := os.ReadDir(dir)
		if len(entries) > 0 {
			foundAt = dir
			foundNonEmtpy = true
			break
		}
	}
	var err error
	if !foundNonEmtpy {
		err = errors.New("non-empty directory not-found")
	}
	return checkResult{
		dependency: dep,
		found:      foundNonEmtpy,
		foundAt:    foundAt,
		err:        err,
	}
}
