package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/concertim/flight-user-suite/flight/doctor"
	"github.com/urfave/cli/v3"
)

func TestWebHelpIncludesDoctor(t *testing.T) {
	cmd := testRootCommand()
	var out bytes.Buffer
	cmd.Writer = &out
	cmd.ErrWriter = &out

	if err := cmd.Run(context.Background(), []string{"flight", "web", "--help"}); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	if !strings.Contains(output, "doctor  System health check for Flight Web Suite dependencies") {
		t.Fatalf("expected web help to include doctor command, got:\n%s", output)
	}
}

func TestWebDoctorHelp(t *testing.T) {
	cmd := testRootCommand()
	var out bytes.Buffer
	cmd.Writer = &out
	cmd.ErrWriter = &out

	if err := cmd.Run(context.Background(), []string{"flight", "web", "doctor", "--help"}); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	if !strings.Contains(output, "flight web doctor [options]") {
		t.Fatalf("expected doctor help usage, got:\n%s", output)
	}
	if !strings.Contains(output, "Perform a health check on the system to check that all Flight Web Suite") ||
		!strings.Contains(output, "dependencies are present.") {
		t.Fatalf("expected doctor help description, got:\n%s", output)
	}
}

func TestWebDoctorSuccess(t *testing.T) {
	deps := setupWebDoctorTestDeps(t, false)
	restoreDeps := swapWebDoctorDependencies(t, deps.general, deps.desktop)
	defer restoreDeps()

	output, err := runWebDoctorForTest(t)
	if err != nil {
		t.Fatalf("expected success, got %v\noutput:\n%s", err, output)
	}
	if !strings.Contains(output, "Required Flight Desktop access dependencies") {
		t.Fatalf("expected desktop dependency section, got:\n%s", output)
	}
}

func TestWebDoctorFailsWhenPythonMissing(t *testing.T) {
	deps := setupWebDoctorTestDeps(t, false)
	deps.general[0].Paths = []string{"missing-python3"}
	deps.general[1].Paths = []string{"missing-python3"}
	restoreDeps := swapWebDoctorDependencies(t, deps.general, deps.desktop)
	defer restoreDeps()

	output, err := runWebDoctorForTest(t)
	assertExitCodeOne(t, err, output)
	if !strings.Contains(output, "Required general Flight Web Suite dependencies not satisfied") {
		t.Fatalf("expected python failure section, got:\n%s", output)
	}
}

func TestWebDoctorFailsWhenPythonPamMissing(t *testing.T) {
	deps := setupWebDoctorTestDeps(t, true)
	restoreDeps := swapWebDoctorDependencies(t, deps.general, deps.desktop)
	defer restoreDeps()

	output, err := runWebDoctorForTest(t)
	assertExitCodeOne(t, err, output)
	if !strings.Contains(output, "Required general Flight Web Suite dependencies not satisfied") {
		t.Fatalf("expected python-pam failure section, got:\n%s", output)
	}
}

func TestWebDoctorFailsWhenWebsockifyMissing(t *testing.T) {
	deps := setupWebDoctorTestDeps(t, false)
	deps.desktop[0].Paths = []string{filepath.Join(t.TempDir(), "missing-websockify")}
	restoreDeps := swapWebDoctorDependencies(t, deps.general, deps.desktop)
	defer restoreDeps()

	output, err := runWebDoctorForTest(t)
	assertExitCodeOne(t, err, output)
	if !strings.Contains(output, "Required Flight Desktop access dependencies not satisfied") {
		t.Fatalf("expected websockify failure section, got:\n%s", output)
	}
}

func TestWebDoctorOptionalImportMissingDoesNotFail(t *testing.T) {
	deps := setupWebDoctorTestDeps(t, false)
	deps.desktop[1].Paths = []string{filepath.Join(t.TempDir(), "missing-import")}
	restoreDeps := swapWebDoctorDependencies(t, deps.general, deps.desktop)
	defer restoreDeps()

	output, err := runWebDoctorForTest(t)
	if err != nil {
		t.Fatalf("expected success with optional dependency missing, got %v\noutput:\n%s", err, output)
	}
	if !strings.Contains(output, "OPTIONAL Flight Desktop access dependencies not satisfied") {
		t.Fatalf("expected optional warning, got:\n%s", output)
	}
}

func testRootCommand() *cli.Command {
	cmd := &cli.Command{
		Name:                   progName,
		Usage:                  "The Flight User Suite",
		Version:                version,
		UseShortOptionHandling: true,
		HideHelpCommand:        true,
	}
	addAdminCommands(cmd, maxTextWidth)
	return cmd
}

type webDoctorTestDeps struct {
	general []doctor.Dependency
	desktop []doctor.Dependency
}

func setupWebDoctorTestDeps(t *testing.T, failPam bool) webDoctorTestDeps {
	t.Helper()

	tmpDir := t.TempDir()
	pythonPath := filepath.Join(tmpDir, "python3")
	websockifyPath := filepath.Join(tmpDir, "websockify")
	importPath := filepath.Join(tmpDir, "import")

	pythonScript := "#!/bin/sh\ncase \"$2\" in\n  \"import pam\") exit 0 ;;\n  *) echo \"unexpected invocation: $*\" >&2; exit 2 ;;\nesac\n"
	if failPam {
		pythonScript = "#!/bin/sh\necho \"ModuleNotFoundError: No module named 'pam'\" >&2\nexit 1\n"
	}
	writeToolFixture(t, pythonPath, 0o755, pythonScript)
	writeToolFixture(t, websockifyPath, 0o755, "#!/bin/sh\nexit 0\n")
	writeToolFixture(t, importPath, 0o755, "#!/bin/sh\nexit 0\n")

	return webDoctorTestDeps{
		general: []doctor.Dependency{
			{
				Type:        doctor.TypeExecutable,
				Description: "Python 3",
				Paths:       []string{"python3"},
			},
			{
				Type:        doctor.TypePythonModule,
				Description: "python-pam",
				Paths:       []string{"python3"},
				Module:      "pam",
			},
		},
		desktop: []doctor.Dependency{
			{
				Type:        doctor.TypeExecutable,
				Description: "Websockify",
				Paths:       []string{websockifyPath},
			},
			{
				Type:           doctor.TypeExecutable,
				Description:    "ImageMagick import",
				Optional:       true,
				FailureMessage: "Screenshot capture for Flight Desktop access will not be available",
				Paths:          []string{importPath},
			},
		},
	}
}

func swapWebDoctorDependencies(t *testing.T, general []doctor.Dependency, desktop []doctor.Dependency) func() {
	t.Helper()
	prevGeneral := webDoctorGeneralDependencies
	prevDesktop := webDoctorDesktopDependencies
	prevDelay := doctor.SpinnerDelay
	prevPath := os.Getenv("PATH")

	pathParts := []string{}
	seen := map[string]bool{}
	for _, dep := range append(append([]doctor.Dependency{}, general...), desktop...) {
		for _, path := range dep.Paths {
			if filepath.IsAbs(path) {
				dir := filepath.Dir(path)
				if !seen[dir] {
					seen[dir] = true
					pathParts = append(pathParts, dir)
				}
			}
		}
	}
	if prevPath != "" {
		pathParts = append(pathParts, prevPath)
	}
	if len(pathParts) > 0 {
		if err := os.Setenv("PATH", strings.Join(pathParts, string(os.PathListSeparator))); err != nil {
			t.Fatal(err)
		}
	}
	doctor.SpinnerDelay = 0
	webDoctorGeneralDependencies = general
	webDoctorDesktopDependencies = desktop

	return func() {
		webDoctorGeneralDependencies = prevGeneral
		webDoctorDesktopDependencies = prevDesktop
		doctor.SpinnerDelay = prevDelay
		_ = os.Setenv("PATH", prevPath)
	}
}

func runWebDoctorForTest(t *testing.T) (string, error) {
	t.Helper()

	stdout := os.Stdout
	stderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	os.Stderr = w
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()

	runErr := runWebDoctor(context.Background(), &cli.Command{})

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.String(), runErr
}

func assertExitCodeOne(t *testing.T, err error, output string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected exit code 1, got success\noutput:\n%s", output)
	}
	var exitErr cli.ExitCoder
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got %v\noutput:\n%s", err, output)
	}
}
