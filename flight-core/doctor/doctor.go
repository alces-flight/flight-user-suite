package doctor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/yarlson/pin"
)

const (
	TypeExecutable = "exe"
	TypeDirectory  = "dir"
)

var SpinnerDelay = 1 * time.Second

type Dependency struct {
	Type           string   `yaml:"type"`
	Description    string   `yaml:"description"`
	Optional       bool     `yaml:"optional"`
	Paths          []string `yaml:"paths"`
	FailureMessage string   `yaml:"failure_message"`
	SuccessMessage string   `yaml:"success_message"`
	Module         string   `yaml:"module"`
}

type CheckResult struct {
	Dependency Dependency
	Found      bool
	FoundAt    string
	Err        error
}

func RequiredDependencies(deps []Dependency) []Dependency {
	req := make([]Dependency, 0, len(deps))
	for _, dep := range deps {
		if !dep.Optional {
			req = append(req, dep)
		}
	}
	return req
}

func OptionalDependencies(deps []Dependency) []Dependency {
	opt := make([]Dependency, 0, len(deps))
	for _, dep := range deps {
		if dep.Optional {
			opt = append(opt, dep)
		}
	}
	return opt
}

func CheckRequiredDeps(ctx context.Context, spinnerText string, doneText string, failText string, deps []Dependency) bool {
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
	timer := time.After(SpinnerDelay)
	checkResults, ok := Run(deps)
	<-timer
	if ok {
		p.Stop(doneText)
	} else {
		p.Fail(failText)
		allOK = false
	}
	PrintCheckResults(checkResults)
	return allOK
}

func CheckOptionalDeps(ctx context.Context, spinnerText string, doneText string, failText string, deps []Dependency) {
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
	timer := time.After(SpinnerDelay)
	checkResults, ok := Run(deps)
	<-timer
	if ok {
		p.Stop(doneText)
	} else {
		p.Fail(failText)
	}
	PrintCheckResults(checkResults)
}

func Run(dependencies []Dependency) ([]CheckResult, bool) {
	checkResults := make([]CheckResult, 0, len(dependencies))
	allOK := true
	for _, dep := range dependencies {
		var result CheckResult
		switch dep.Type {
		case TypeExecutable:
			result = checkExeAvailable(dep)
		case TypeDirectory:
			result = checkDirNonEmpty(dep)
		default:
			result = CheckResult{
				Dependency: dep,
				Err:        fmt.Errorf("unknown dependency type: %s", dep.Type),
			}
		}
		if result.Err != nil {
			allOK = false
		}
		checkResults = append(checkResults, result)
	}
	return checkResults, allOK
}

func PrintCheckResults(checkResults []CheckResult) {
	depParts := make([]string, 0, len(checkResults))
	resultParts := make([]string, 0, len(checkResults))
	for _, result := range checkResults {
		tick := " > \u2705 "
		outcome := result.FoundAt
		if len(result.Dependency.SuccessMessage) > 0 {
			outcome = lipgloss.Wrap(result.Dependency.SuccessMessage, 60, "")
		}
		if !result.Found {
			tick = " > \u274c "
			if len(result.Dependency.FailureMessage) > 0 {
				outcome = lipgloss.Wrap(result.Dependency.FailureMessage, 60, "")
			} else if result.Err != nil {
				outcome = result.Err.Error()
			}
		}
		description := result.Dependency.Description
		if description == "" {
			description = strings.Join(result.Dependency.Paths, "\n")
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

func checkExeAvailable(dep Dependency) CheckResult {
	var errs error
	for _, path := range dep.Paths {
		location, err := exec.LookPath(path)
		if err == nil {
			return CheckResult{
				Dependency: dep,
				Found:      true,
				FoundAt:    location,
			}
		}
		errs = errors.Join(errs, err)
	}
	return CheckResult{
		Dependency: dep,
		Err:        errs,
	}
}

func checkDirNonEmpty(dep Dependency) CheckResult {
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
		return CheckResult{
			Dependency: dep,
			Found:      true,
			FoundAt:    strings.Join(nonEmptyDirs, "\n"),
		}
	}
	return CheckResult{
		Dependency: dep,
		Err:        errs,
	}
}
