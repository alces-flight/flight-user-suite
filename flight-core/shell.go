package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"charm.land/log/v2"
	"github.com/adrg/xdg"
	"github.com/concertim/flight-user-suite/flight/cliui"
	"github.com/concertim/flight-user-suite/flight/toolset"
	"github.com/ergochat/readline"
	"github.com/urfave/cli/v3"
)

// Return list of tool names, if the first word in the line is not yet
// complete.
func toolCompletions(line string) []string {
	// The command is complete, if there is more than one word on the line, or
	// there is a single word followed by a space.
	var cmdComplete bool
	if strings.TrimSpace(line) == "" {
		cmdComplete = false
		// TODO: Support quotes. AKA shellwords.
	} else if len(strings.Split(strings.TrimSpace(line), " ")) > 1 {
		cmdComplete = true
	} else if strings.HasSuffix(line, " ") {
		cmdComplete = true
	}

	tools, err := toolset.GetTools(env.FlightRoot, false)
	if err != nil {
		log.Warn("Error", "err", err)
		return nil
	}
	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		toolNames = append(toolNames, tool.Name)
	}

	if !cmdComplete {
		return toolNames
	}

	finalWordComplete := strings.HasSuffix(line, " ")
	line = strings.TrimSpace(line)
	// TODO: Support quotes. AKA shellwords.
	words := strings.Split(line, " ")
	cmd := words[0]

	if !slices.Contains(toolNames, cmd) {
		// The commmand is not one of the tools (perhaps its a builtin such as
		// "help"), we offer no completions.
		return nil
	}
	tool := &toolset.Tool{Name: cmd}

	// Request completion from tool. Omit the final word if incomplete.
	var toolArgs []string
	if finalWordComplete {
		toolArgs = make([]string, len(words)-1)
	} else {
		toolArgs = make([]string, len(words)-2)
	}
	copy(toolArgs, words[1:])
	toolArgs = append(toolArgs, "--generate-shell-completion")
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	bytes, err := runToolWithOutput(ctx, tool, toolArgs)
	out := string(bytes)
	if err != nil {
		log.Warn("Error", "err", err)
		return nil
	}

	lastWord := words[len(words)-1]
	comps := make([]string, 0)
	for comp := range strings.SplitSeq(out, "\n") {
		if comp == "" {
			continue
		}
		// Filter the completions offered by the tool. If the final word of the
		// line is complete, we want all completions offered; otherwise, we
		// want only those with a prefix matching the incomplete last word.
		if finalWordComplete || strings.HasPrefix(comp, lastWord) {
			var prefix string
			if finalWordComplete {
				prefix = line
			} else if len(words) >= 2 {
				prefix = strings.Join(words[0:len(words)-1], " ")
			}
			comps = append(comps, fmt.Sprintf("%s %s", prefix, comp))
		}
	}
	return comps
}

var completer = readline.NewPrefixCompleter(
	readline.PcItemDynamic(toolCompletions),
	readline.PcItem("help"),
	readline.PcItem("clear"),
	readline.PcItem("exit"),
)

func shellUsage(w io.Writer) {
	io.WriteString(w, "Available Flight tools:\n")
	tools := toolCompletions("")
	for _, tool := range tools {
		// TODO: Add tool synopsis here.
		fmt.Fprintf(w, "    %s\n", tool)
	}
	io.WriteString(w, "Shell builtin commands:\n")
	// TODO: Can we add a synopsis here?
	io.WriteString(w, completer.Tree("    "))
}

func execInput(baseTool, input string, rl *readline.Instance) error {
	input = strings.TrimSpace(input)
	// TODO: Support quotes. AKA shellwords.
	args := strings.Split(input, " ")

	if baseTool != "" {
		if args[0] == "" {
			return nil
		} else {
			return shellRunTool(baseTool, args)
		}
	} else {
		// Check for built-in commands.
		switch args[0] {
		case "":
			return nil
		case "help":
			shellUsage(rl.Stderr())
			return nil
		case "clear":
			rl.ClearScreen()
			return nil
		case "exit":
			os.Exit(0)
		}
		tool := args[0]
		args = args[1:]
		return shellRunTool(tool, args)
	}
}

func shellRunTool(tool string, args []string) error {
	tp := toolPath(tool)
	cmd := exec.Command(tp, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	return transformToolError(tool, err)
}

func runShell(ctx context.Context, cmd *cli.Command) error {
	baseTool := cmd.StringArg("tool")

	prompt := cliui.PromptStyle.Render("flight» ")
	if baseTool != "" {
		if !slices.Contains(toolCompletions(""), baseTool) {
			return UnknownTool{Tool: baseTool}
		}
		prompt = cliui.PromptStyle.Render(fmt.Sprintf("flight %s» ", baseTool))
	}
	historyFile, err := xdg.CacheFile(filepath.Join("flight", "shell", "history"))
	if err != nil {
		historyFile = ""
	}
	rlConfig := &readline.Config{
		Prompt:            prompt,
		HistoryFile:       historyFile,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		Undo:              true,
	}
	if baseTool == "" {
		rlConfig.AutoComplete = completer
	}
	rl, err := readline.NewFromConfig(rlConfig)
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
		if err = execInput(baseTool, line, rl); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	return nil
}
