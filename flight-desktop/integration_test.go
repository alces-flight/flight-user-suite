package main_test

import (
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
