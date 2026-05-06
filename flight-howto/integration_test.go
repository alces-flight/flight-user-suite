package main_test

import (
	"encoding/json"
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
var flightStateRoot = ""

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
	flightStateRoot = filepath.Join(tmpDir, "state")
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
		skipCI           bool
	}{
		{
			"--help outputs expected help",
			[]string{"--help"},
			"golden/help.golden",
			0,
			false,
		},
		{
			"list --help outputs expected help",
			[]string{"list", "--help"},
			"golden/list-help.golden",
			0,
			false,
		},
		{
			"show --help outputs expected help",
			[]string{"show", "--help"},
			"golden/show-help.golden",
			0,
			false,
		},
		{
			"list shows expected table when there are howto guides",
			[]string{"list"},
			"golden/list-non-empty.golden",
			0,
			true,
		},
		{
			"show displays error message when index is not known",
			[]string{"show", "0"},
			"golden/show-bad-index.golden",
			1,
			false,
		},
		{
			"show displays file contents when index is good",
			[]string{"show", "1"},
			"golden/show-guide-1.golden",
			0,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if tt.skipCI {
				t.Skip("test needs updating to work for admin users")
			}
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

func Test_list_json(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("test needs updating to work for admin users")
	}
	output, err := runBinary([]string{"list", "--format", "json"}, nil)
	assertExitCode(t, 0, output, err)

	var response struct {
		Success bool `json:"success"`
		Guides  []struct {
			Index int    `json:"index"`
			Title string `json:"title"`
		} `json:"guides"`
		Error  string `json:"error"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(output, &response); err != nil {
		t.Fatalf("unmarshal output: %v\noutput:\n%s", err, output)
	}

	if !response.Success {
		t.Fatalf("expected success response: %#v", response)
	}
	if response.Error != "" || response.Reason != "" {
		t.Fatalf("unexpected error fields: %#v", response)
	}

	guides := response.Guides
	if len(guides) != 3 {
		t.Fatalf("unexpected guides: %#v", guides)
	}
	if guides[0].Index != 1 || guides[0].Title != "Base Guide A" {
		t.Fatalf("unexpected first guide: %#v", guides)
	}
	if guides[1].Index != 2 || guides[1].Title != "Base Guide B" {
		t.Fatalf("unexpected second guide: %#v", guides)
	}
	if guides[2].Index != 3 || guides[2].Title != "My Category > Guide" {
		t.Fatalf("unexpected third guide: %#v", guides)
	}

	for _, guide := range guides {
		if guide.Title == "Admin Guides > Admin Guide" {
			t.Fatalf("admin-only guide should be filtered from json output: %#v", guides)
		}
	}
}

func Test_show_json_success(t *testing.T) {
	output, err := runBinary([]string{"show", "--format", "json", "1"}, nil)
	assertExitCode(t, 0, output, err)

	var response struct {
		Success bool `json:"success"`
		Guide   struct {
			Title       string `json:"title"`
			RawMarkdown string `json:"raw_markdown"`
		} `json:"guide"`
	}
	if err := json.Unmarshal(output, &response); err != nil {
		t.Fatalf("unmarshal output: %v\noutput:\n%s", err, output)
	}

	if !response.Success {
		t.Fatalf("expected success response: %#v", response)
	}
	if response.Guide.Title != "Base Guide A" {
		t.Fatalf("unexpected title: %#v", response)
	}
	if response.Guide.RawMarkdown != "Base guide A\n" {
		t.Fatalf("unexpected raw markdown: %#v", response)
	}
}

func Test_show_json_invalid_index(t *testing.T) {
	output, err := runBinary([]string{"show", "--format", "json", "0"}, nil)
	assertExitCode(t, 1, output, err)

	var response struct {
		Success bool `json:"success"`
		Guide   struct {
			Title       string `json:"title"`
			RawMarkdown string `json:"raw_markdown"`
		} `json:"guide"`
		Error  string `json:"error"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(extractJSONOutput(output), &response); err != nil {
		t.Fatalf("unmarshal output: %v\noutput:\n%s", err, output)
	}

	if response.Success {
		t.Fatalf("expected failure response: %#v", response)
	}
	if response.Reason != "invalid_index" {
		t.Fatalf("unexpected reason: %#v", response)
	}
	if response.Guide != (struct {
		Title       string "json:\"title\""
		RawMarkdown string "json:\"raw_markdown\""
	}{}) {
		t.Fatalf("expected empty guide on failure: %#v", response)
	}
}

func Test_show_json_missing_index(t *testing.T) {
	output, err := runBinary([]string{"show", "--format", "json"}, nil)
	assertExitCode(t, 1, output, err)

	var response struct {
		Success bool `json:"success"`
		Guide   struct {
			Title       string `json:"title"`
			RawMarkdown string `json:"raw_markdown"`
		} `json:"guide"`
		Error  string `json:"error"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(extractJSONOutput(output), &response); err != nil {
		t.Fatalf("unmarshal output: %v\noutput:\n%s", err, output)
	}

	if response.Success {
		t.Fatalf("expected failure response: %#v", response)
	}
	if response.Reason != "unexpected" {
		t.Fatalf("unexpected reason: %#v", response)
	}
	if response.Error != "Incorrect usage: missing argument index" {
		t.Fatalf("unexpected error: %#v", response)
	}
}

func Test_show_json_invalid_index_flag_after_arg(t *testing.T) {
	output, err := runBinary([]string{"show", "-1", "--format", "json"}, nil)
	assertExitCode(t, 1, output, err)

	var response struct {
		Success bool `json:"success"`
		Guide   struct {
			Title       string `json:"title"`
			RawMarkdown string `json:"raw_markdown"`
		} `json:"guide"`
		Error  string `json:"error"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(extractJSONOutput(output), &response); err != nil {
		t.Fatalf("unmarshal output: %v\noutput:\n%s", err, output)
	}

	if response.Success {
		t.Fatalf("expected failure response: %#v", response)
	}
	if response.Reason != "invalid_index" {
		t.Fatalf("unexpected reason: %#v", response)
	}
}

func Test_show_json_not_found(t *testing.T) {
	linkPath := filepath.Join(flightRoot, "usr", "share", "doc", "howtos-enabled", "01-base-guide-a.md")
	if err := os.Remove(linkPath); err != nil {
		t.Fatal(err)
	}
	originalPath := filepath.Join(testdataPath, "howtos", "01-base-guide-a.md")
	t.Cleanup(func() {
		_ = os.Remove(linkPath)
		if err := os.Symlink(originalPath, linkPath); err != nil {
			t.Fatalf("restore guide symlink: %v", err)
		}
	})
	if err := os.Symlink(filepath.Join(tmpDir, "missing-guide.md"), linkPath); err != nil {
		t.Fatal(err)
	}

	output, err := runBinary([]string{"show", "--format", "json", "1"}, nil)
	assertExitCode(t, 1, output, err)

	var response struct {
		Success bool `json:"success"`
		Guide   struct {
			Title       string `json:"title"`
			RawMarkdown string `json:"raw_markdown"`
		} `json:"guide"`
		Error  string `json:"error"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(extractJSONOutput(output), &response); err != nil {
		t.Fatalf("unmarshal output: %v\noutput:\n%s", err, output)
	}

	if response.Success {
		t.Fatalf("expected failure response: %#v", response)
	}
	if response.Reason != "not_found" {
		t.Fatalf("unexpected reason: %#v", response)
	}
	if response.Error != "Unknown howto: 01-base-guide-a.md" {
		t.Fatalf("unexpected error: %#v", response)
	}
	if response.Guide != (struct {
		Title       string "json:\"title\""
		RawMarkdown string "json:\"raw_markdown\""
	}{}) {
		t.Fatalf("expected empty guide on failure: %#v", response)
	}
}

func Test_show_json_collecting_guides_error(t *testing.T) {
	missingRoot := filepath.Join(tmpDir, "missing-root")
	output, err := runBinaryWithEnv(
		[]string{"show", "-1", "--format", "json"},
		nil,
		[]string{fmt.Sprintf("FLIGHT_ROOT=%s", missingRoot)},
	)
	assertExitCode(t, 1, output, err)

	var response struct {
		Success bool `json:"success"`
		Guide   struct {
			Title       string `json:"title"`
			RawMarkdown string `json:"raw_markdown"`
		} `json:"guide"`
		Error  string `json:"error"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(extractJSONOutput(output), &response); err != nil {
		t.Fatalf("unmarshal output: %v\noutput:\n%s", err, output)
	}

	if response.Success {
		t.Fatalf("expected failure response: %#v", response)
	}
	if response.Reason != "unexpected" {
		t.Fatalf("unexpected reason: %#v", response)
	}
	expectedError := fmt.Sprintf(
		"collecting guide files: reading directory: open %s/usr/share/doc/howtos-enabled: no such file or directory",
		missingRoot,
	)
	if response.Error != expectedError {
		t.Fatalf("unexpected error: %#v", response)
	}
}

func Test_list_json_load_error(t *testing.T) {
	missingRoot := filepath.Join(tmpDir, "missing-root")
	output, err := runBinaryWithEnv(
		[]string{"list", "--format", "json"},
		nil,
		[]string{fmt.Sprintf("FLIGHT_ROOT=%s", missingRoot)},
	)
	assertExitCode(t, 1, output, err)

	var response struct {
		Success bool `json:"success"`
		Guides  []struct {
			Index int    `json:"index"`
			Title string `json:"title"`
		} `json:"guides"`
		Error  string `json:"error"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(extractJSONOutput(output), &response); err != nil {
		t.Fatalf("unmarshal output: %v\noutput:\n%s", err, output)
	}

	if response.Success {
		t.Fatalf("expected failure response: %#v", response)
	}
	if response.Reason != "unexpected" {
		t.Fatalf("unexpected reason: %#v", response)
	}
	if len(response.Guides) != 0 {
		t.Fatalf("expected empty guides on failure: %#v", response)
	}
	expectedError := fmt.Sprintf(
		"reading directory: open %s/usr/share/doc/howtos-enabled: no such file or directory",
		missingRoot,
	)
	if response.Error != expectedError {
		t.Fatalf("unexpected error: %#v", response)
	}
}

func extractJSONOutput(output []byte) []byte {
	return []byte(strings.TrimSuffix(string(output), "exit status 1\n"))
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
	cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata", fmt.Sprintf("FLIGHT_ROOT=%s", flightRoot), fmt.Sprintf("FLIGHT_STATE_ROOT=%s", flightStateRoot))
	if stdin != nil {
		cmd.Stdin = strings.NewReader(*stdin)
	}
	return cmd.CombinedOutput()
}

func runBinaryWithEnv(args []string, stdin *string, extraEnv []string) ([]byte, error) {
	fullArgs := append([]string{"run", entryPoint}, args...)
	cmd := exec.Command("go", fullArgs...)
	cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata", fmt.Sprintf("FLIGHT_ROOT=%s", flightRoot), fmt.Sprintf("FLIGHT_STATE_ROOT=%s", flightStateRoot))
	cmd.Env = append(cmd.Env, extraEnv...)
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
