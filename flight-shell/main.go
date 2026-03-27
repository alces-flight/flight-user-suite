package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	flightRoot string = "/opt/flight"
	toolDir    string
)

func init() {
	if root, ok := os.LookupEnv("FLIGHT_ROOT"); ok {
		flightRoot = root
	}
	toolDir = filepath.Join(flightRoot, "usr", "lib", "flight-core")
}

func toolPath(tool string) string {
	return filepath.Join(toolDir, fmt.Sprintf("flight-%s", tool))
}

type DisabledTool struct {
	Tool string
}

func (ut DisabledTool) Error() string {
	return fmt.Sprintf("The %s tool is not enabled", ut.Tool)
}

type UnknownTool struct {
	Tool string
}

func (ut UnknownTool) Error() string {
	return fmt.Sprintf("Unknown tool: %s", ut.Tool)
}

func transformToolError(tool string, err error) error {
	if pathError, ok := errors.AsType[*fs.PathError](err); ok {
		if pathError.Err.Error() == "no such file or directory" {
			return UnknownTool{Tool: tool}
		}
		if pathError.Err.Error() == "permission denied" {
			return DisabledTool{Tool: tool}
		}
	}
	return err
}

func execInput(input string) error {
	input = strings.TrimSpace(input)

	// TODO: Support quotes.
	args := strings.Split(input, " ")

	// Check for built-in commands.
	// TODO: Add `help` to display list of commands and synopsis for them.
	// TODO: Do we want to support the `tools` and `hooks` commands?  What's the use case for this shell?
	switch args[0] {
	case "":
		return nil
	case "exit":
		os.Exit(0)
	}

	tool := args[0]
	tp := toolPath(tool)
	cmd := exec.Command(tp, args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	return transformToolError(tool, err)
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("flight> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println()
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "\nerror reading input: %s\n", err)
		}
		if err = execInput(input); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
