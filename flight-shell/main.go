package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ergochat/readline"
)

var (
	flightRoot  string = "/opt/flight"
	toolDir     string
	ctmOrange   = lipgloss.Color("#ff7401")
	promptStyle = lipgloss.NewStyle().Foreground(ctmOrange)
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

func getTools(line string) []string {
	// Remove duplication with flight-core#getTools.
	entries, _ := os.ReadDir(toolDir)
	tools := make([]string, 0)
	for _, entry := range entries {
		if tool, hasPrefix := strings.CutPrefix(entry.Name(), "flight-"); hasPrefix {
			info, _ := entry.Info()
			if info.Mode()&0111 != 0 {
				tools = append(tools, tool)
			}
		}
	}
	return tools
}

var completer = readline.NewPrefixCompleter(
	readline.PcItemDynamic(getTools),
	readline.PcItem("help"),
	readline.PcItem("clear"),
)

func help(w io.Writer) {
	io.WriteString(w, "commands:\n")
	tools := getTools("")
	for _, tool := range tools {
		fmt.Fprintf(w, "    %s\n", tool)
	}
	io.WriteString(w, completer.Tree("    "))
}

func execInput(input string, rl *readline.Instance) error {
	input = strings.TrimSpace(input)

	// TODO: Support quotes.
	args := strings.Split(input, " ")

	// Check for built-in commands.
	// TODO: Add `help` to display list of commands and synopsis for them.
	// TODO: Do we want to support the `tools` and `hooks` commands?  What's the use case for this shell?
	switch args[0] {
	case "":
		return nil
	case "help":
		help(rl.Stderr())
		return nil
	case "clear":
		rl.ClearScreen()
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
	rl, err := readline.NewFromConfig(&readline.Config{
		Prompt: promptStyle.Render("flight» "),
		// TODO: Use an XDG cache/data dir for history file.
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold: true,

		Undo: true,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()
	rl.CaptureExitSignal()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "\nerror reading input: %s\n", err)
		}
		if err = execInput(line, rl); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
