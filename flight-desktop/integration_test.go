package main_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var update = flag.Bool("update", false, "update golden files")
var keepTmpDir = flag.Bool("keep-tmpdir", false, "keep temporary director for later inspection")

var entryPoint = "./..."
var goRoot = ""
var testdataPath = ""
var tmpDir = ""
var flightRoot = ""
var flightStateRoot = ""

type availableTypeResponse struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	URL     string `json:"url"`
}

type startCommandResponse struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error"`
	Reason      string `json:"reason"`
}

// Setup/teardown logic for running all tests in the package.
func TestMain(m *testing.M) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("problems recovering caller information")
		os.Exit(1)
	}
	goRoot = filepath.Dir(filename)
	testdataPath = filepath.Join(goRoot, "testdata")
	tmpDir = createTempDir("flight-desktop-")
	fmt.Printf("tmpDir: %v\n", tmpDir)
	flightRoot = filepath.Join(tmpDir, "opt", "flight")
	flightStateRoot = filepath.Join(tmpDir, "state")
	installTestFiles()
	exitCode := m.Run()
	if !*keepTmpDir {
		_ = os.RemoveAll(tmpDir)
	}
	os.Exit(exitCode)
}

func Test_golden_tests(t *testing.T) {
	tests := []struct {
		testName         string
		optionsAndArgs   []string
		fixture          string
		expectedExitCode int
	}{
		{
			"--help outputs expected help",
			[]string{"--help"},
			"golden/help.golden",
			0,
		},
		{
			"avail --help outputs expected help",
			[]string{"avail", "--help"},
			"golden/avail-help.golden",
			0,
		},
		{
			"doctor --help outputs expected help",
			[]string{"doctor", "--help"},
			"golden/doctor-help.golden",
			0,
		},
		{
			"start --help outputs expected help",
			[]string{"start", "--help"},
			"golden/start-help.golden",
			0,
		},
		{
			"list --help outputs expected help",
			[]string{"list", "--help"},
			"golden/list-help.golden",
			0,
		},
		{
			"show --help outputs expected help",
			[]string{"show", "--help"},
			"golden/show-help.golden",
			0,
		},
		{
			"kill --help outputs expected help",
			[]string{"kill", "--help"},
			"golden/kill-help.golden",
			0,
		},
		{
			"clean --help outputs expected help",
			[]string{"clean", "--help"},
			"golden/clean-help.golden",
			0,
		},
		{
			"avail shows expected table",
			[]string{"avail"},
			"golden/avail.golden",
			0,
		},
		{
			"doctor shows expected output",
			[]string{"doctor"},
			"golden/doctor.golden",
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			output, err := runBinary(tt.optionsAndArgs, nil)
			assertExitCode(t, tt.expectedExitCode, output, err)
			if *update {
				writeFixture(t, tt.fixture, output)
			}
			expected := loadFixture(t, tt.fixture)
			assertOutput(t, expected, output)
		})
	}
}

func Test_session_start_artefacts(t *testing.T) {
	args := []string{"start", "xterm", "--name", "my-session"}
	output, err := runBinary(args, nil)
	assertExitCode(t, 0, output, err)
	sessionDir := filepath.Join(tmpDir, "local", "state", "flight", "desktop", "sessions", "my-session")
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		t.Fatal(err)
	}
	fileNames := make([]string, 0)
	for _, entry := range entries {
		fileNames = append(fileNames, entry.Name())
	}
	// Only check the files that are created by our Go program.  Those created
	// by the libexec/desktop/vncserver script are not checked.
	expectedFileNames := []string{"metadata.yml", "session.sh"}
	for _, efn := range expectedFileNames {
		if !slices.Contains(fileNames, efn) {
			t.Logf("expected file '%s' to be in session dir %s, but session dir contains only\n  %s", efn, sessionDir, fileNames)
			t.Fail()
		}
	}
	// Clean up after ourselves so other tests have a clean slate.
	args = []string{"clean", "my-session"}
	_, err = runBinary(args, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_session_life_cycle(t *testing.T) {
	tests := []struct {
		testName         string
		optionsAndArgs   []string
		expectedLines    []string
		expectedExitCode int
	}{
		{
			"list shows no sessions",
			[]string{"list"},
			[]string{"No desktop sessions found."},
			0,
		},
		{
			"start a session",
			[]string{"start", "xterm", "--name", "my-session"},
			[]string{
				"Starting 'xterm' desktop session 'my-session':",
				"Checking system dependencies...",
				"Dependencies OK",
				"Starting session...",
				"Your xterm session is ready!",
				"* Primary:        127.0.0.1:5901",
			},
			0,
		},
		{
			"list shows my session",
			[]string{"list"},
			[]string{
				"my-session │ xterm  │ 127.0.0.1:5901",
			},
			0,
		},
		{
			"show my session",
			[]string{"show", "my-session"},
			[]string{
				"Name  my-session",
				"Type  xterm",
				// It's exited 'cos we don't actually create a VNC process
				// during the test.
				"State  exited",
			},
			0,
		},
		{
			"clean my session",
			[]string{"clean", "my-session"},
			[]string{
				"Cleaning desktop sessions",
				"Cleaning complete",
			},
			0,
		},
		{
			"list shows no sessions",
			[]string{"list"},
			[]string{"No desktop sessions found."},
			0,
		},
	}

	for _, tt := range tests {
		output, err := runBinary(tt.optionsAndArgs, nil)
		assertExitCode(t, tt.expectedExitCode, output, err)
		for _, line := range tt.expectedLines {
			assertOutputContains(t, line, output)
		}
	}
}

func Test_avail_json(t *testing.T) {
	output, err := runBinary([]string{"avail", "--format", "json"}, nil)
	assertExitCode(t, 0, output, err)

	var response []availableTypeResponse
	if err := json.Unmarshal(jsonPayload(t, output), &response); err != nil {
		t.Fatalf("failed to decode avail json: %v\noutput:\n%s", err, output)
	}

	if len(response) != 2 {
		t.Fatalf("expected 2 desktop types, got %d", len(response))
	}
	if response[0].ID != "gnome" || response[1].ID != "xterm" {
		t.Fatalf("expected types to be ordered by id, got %#v", response)
	}
}

func Test_avail_json_empty(t *testing.T) {
	typesDir := filepath.Join(flightRoot, "usr", "lib", "desktop", "types")
	backupDir := filepath.Join(flightRoot, "usr", "lib", "desktop", "types-backup")
	if err := os.Rename(typesDir, backupDir); err != nil {
		t.Fatalf("failed to move types dir aside: %v", err)
	}
	if err := os.MkdirAll(typesDir, 0o755); err != nil {
		t.Fatalf("failed to create empty types dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(typesDir)
		_ = os.Rename(backupDir, typesDir)
	})

	output, err := runBinary([]string{"avail", "--format", "json"}, nil)
	assertExitCode(t, 0, output, err)

	var response []availableTypeResponse
	if err := json.Unmarshal(jsonPayload(t, output), &response); err != nil {
		t.Fatalf("failed to decode avail json: %v\noutput:\n%s", err, output)
	}
	if len(response) != 0 {
		t.Fatalf("expected empty desktop type list, got %#v", response)
	}
}

func Test_start_json(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		setup           func(t *testing.T)
		wantExitCode    int
		wantSuccess     bool
		wantSessionName string
		wantReason      string
		wantNamePrefix  string // We're using the "meaningless" name generator so prefix is correct.
	}{
		{
			name:            "success with explicit name",
			args:            []string{"start", "xterm", "--name", "my-session", "--format", "json"},
			wantExitCode:    0,
			wantSuccess:     true,
			wantSessionName: "my-session",
		},
		{
			name:           "success with generated name",
			args:           []string{"start", "xterm", "--format", "json"},
			wantExitCode:   0,
			wantSuccess:    true,
			wantNamePrefix: "xterm.",
		},
		{
			name:            "invalid type",
			args:            []string{"start", "missing", "--format", "json"},
			wantExitCode:    1,
			wantSuccess:     false,
			wantSessionName: "",
			wantReason:      "invalid_type",
		},
		{
			name:            "invalid name",
			args:            []string{"start", "xterm", "--name", "bad name", "--format", "json"},
			wantExitCode:    1,
			wantSuccess:     false,
			wantSessionName: "bad name",
			wantReason:      "invalid_name",
		},
		{
			name:            "invalid geometry",
			args:            []string{"start", "xterm", "--geometry", "1024", "--name", "bad-geometry", "--format", "json"},
			wantExitCode:    1,
			wantSuccess:     false,
			wantSessionName: "bad-geometry",
			wantReason:      "invalid_geometry",
		},
		{
			name:            "dependency failure",
			args:            []string{"start", "gnome", "--name", "my-gnome", "--format", "json"},
			wantExitCode:    1,
			wantSuccess:     false,
			wantSessionName: "my-gnome",
			wantReason:      "dependencies_failed",
		},
		{
			name: "session start failure",
			args: []string{"start", "xterm", "--name", "broken-session", "--format", "json"},
			setup: func(t *testing.T) {
				path := filepath.Join(flightRoot, "usr", "libexec", "desktop", "vncserver")
				original, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read vncserver fixture: %v", err)
				}
				replacement := []byte("#!/bin/sh\nexit 1\n")
				if err := os.WriteFile(path, replacement, 0o755); err != nil {
					t.Fatalf("failed to replace vncserver fixture: %v", err)
				}
				t.Cleanup(func() {
					_ = os.WriteFile(path, original, 0o755)
				})
			},
			wantExitCode:    1,
			wantSuccess:     false,
			wantSessionName: "broken-session",
			wantReason:      "start_failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			output, err := runBinary(tt.args, nil)
			assertExitCode(t, tt.wantExitCode, output, err)

			var response startCommandResponse
			if err := json.Unmarshal(jsonPayload(t, output), &response); err != nil {
				t.Fatalf("failed to decode start json: %v\noutput:\n%s", err, output)
			}

			if response.Success != tt.wantSuccess {
				t.Fatalf("expected success=%t, got %#v", tt.wantSuccess, response)
			}
			if tt.wantSessionName != "" || response.SessionName == "" {
				if response.SessionName != tt.wantSessionName {
					t.Fatalf("expected session name %q, got %#v", tt.wantSessionName, response)
				}
			}
			if tt.wantNamePrefix != "" && !strings.HasPrefix(response.SessionName, tt.wantNamePrefix) {
				t.Fatalf("expected session name to start with %q, got %#v", tt.wantNamePrefix, response)
			}
			if response.Reason != tt.wantReason {
				t.Fatalf("expected reason %q, got %#v", tt.wantReason, response)
			}
		})
	}
}

func fixturePath(fixture string) string {
	return filepath.Join(testdataPath, fixture)
}

func writeFixture(t *testing.T, fixture string, content []byte) {
	err := os.WriteFile(fixturePath(fixture), content, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func loadFixture(t *testing.T, fixture string) string {
	content, err := os.ReadFile(fixturePath(fixture))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

func runBinary(args []string, stdin *string) ([]byte, error) {
	fullArgs := append([]string{"run", entryPoint}, args...)
	cmd := exec.Command("go", fullArgs...)
	cmd.Env = append(
		os.Environ(),
		"GOCOVERDIR=.coverdata",
		fmt.Sprintf("FLIGHT_ROOT=%s", flightRoot),
		fmt.Sprintf("FLIGHT_STATE_ROOT=%s", flightStateRoot),
		fmt.Sprintf("XDG_STATE_HOME=%s", filepath.Join(tmpDir, "local", "state")),
	)
	if stdin != nil {
		cmd.Stdin = strings.NewReader(*stdin)
	}
	return cmd.CombinedOutput()
}

func assertExitCode(t *testing.T, expectedExitCode int, output []byte, err error) {
	t.Helper()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != expectedExitCode {
				t.Fatalf("output:\n%s\nerror:\n%s\n", output, err)
			}
		} else {
			t.Fatalf("output:\n%s\nerror:\n%s\n", output, err)
		}
	}
}

func assertOutput(t *testing.T, expectedOutput string, output []byte) {
	t.Helper()
	actual := string(output)
	expected := expectedOutput
	if !reflect.DeepEqual(actual, expected) {
		t.Logf("expected\n%s\n  got\n%s", expected, actual)
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(expected, actual, false)
		t.Log(dmp.DiffPrettyText(diffs))
		t.FailNow()
	}
}

func assertOutputContains(t *testing.T, expectedLine string, output []byte) {
	t.Helper()
	actual := string(output)

	if !strings.Contains(actual, expectedLine) {
		t.Logf("expected output to contain '%s'\n  got\n%s", expectedLine, actual)
		t.FailNow()
	}
}

func jsonPayload(t *testing.T, output []byte) []byte {
	t.Helper()

	trimmed := bytes.TrimSpace(output)
	if len(trimmed) == 0 {
		t.Fatal("expected JSON output, got empty output")
	}

	lastObject := bytes.LastIndexByte(trimmed, '}')
	lastArray := bytes.LastIndexByte(trimmed, ']')
	last := max(lastObject, lastArray)
	if last == -1 {
		t.Fatalf("expected JSON payload, got %q", trimmed)
	}
	return trimmed[:last+1]
}

func createTempDir(prefix string) string {
	path, err := os.MkdirTemp("", prefix)
	panicIfErr(err)
	return path
}

func installTestFiles() {
	dst := filepath.Join(tmpDir)
	src := filepath.Join(goRoot, "testdata", "setup")
	err := os.CopyFS(dst, os.DirFS(src))
	panicIfErr(err)
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
