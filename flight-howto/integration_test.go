package main_test

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
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
	tmpDir = createTempDir("flight-howto-")
	fmt.Printf("tmpDir: %v\n", tmpDir)
	flightRoot = filepath.Join(tmpDir, "opt", "flight")
	symlinkTestHowtos()
	installThemeFile()

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
			"list shows expected table when there are howto guides",
			[]string{"list"},
			"golden/list-non-empty.golden",
			0,
		},
		{
			"show displays error message when index is not known",
			[]string{"show", "0"},
			"golden/show-bad-index.golden",
			1,
		},
		{
			"show displays file contents when index is good",
			[]string{"show", "1"},
			"golden/show-guide-1.golden",
			0,
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
	cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata", fmt.Sprintf("FLIGHT_ROOT=%s", flightRoot))
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

func createTempDir(prefix string) string {
	path, err := os.MkdirTemp("", prefix)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(filepath.Join(path, "opt", "flight", "usr", "share", "doc", "howtos-enabled"), 0o755)
	if err != nil {
		panic(err)
	}
	return path
}

func symlinkTestHowtos() {
	baseDir := filepath.Join(testdataPath, "howtos")
	howtoDir := filepath.Join(flightRoot, "usr", "share", "doc", "howtos-enabled")

	matches, err := filepath.Glob(filepath.Join(baseDir, "*.md"))
	panicIfErr(err)
	matchesDeep, err := filepath.Glob(filepath.Join(baseDir, "*", "*.md"))
	matches = append(matches, matchesDeep...)
	panicIfErr(err)
	for _, match := range matches {
		rel, err := filepath.Rel(baseDir, match)
		panicIfErr(err)
		relDir := filepath.Dir(rel)
		err = os.MkdirAll(filepath.Join(howtoDir, relDir), 0o755)
		panicIfErr(err)
		linkName := filepath.Join(howtoDir, rel)
		err = os.Symlink(match, linkName)
		panicIfErr(err)
	}
}

func installThemeFile() {
	dst := filepath.Join(flightRoot, "usr", "lib", "flight-howto", "themes")
	src := filepath.Join(goRoot, "opt", "flight", "usr", "lib", "flight-howto", "themes")
	err := os.CopyFS(dst, os.DirFS(src))
	panicIfErr(err)
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
