package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"charm.land/log/v2"
	"github.com/ergochat/readline"
	"github.com/urfave/cli/v3"
)

func shellGetTools(line string) []string {
	tools, err := getTools(true)
	if err != nil {
		log.Warn("Error", "err", err)
		return nil
	}
	return tools
}

var completer = readline.NewPrefixCompleter(
	readline.PcItemDynamic(shellGetTools),
	readline.PcItem("help"),
	readline.PcItem("clear"),
	readline.PcItem("exit"),
)

func shellUsage(w io.Writer) {
	io.WriteString(w, "Available Flight tools:\n")
	tools := shellGetTools("")
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
	// TODO: Support quotes.
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

	prompt := promptStyle.Render("flight» ")
	if baseTool != "" {
		if !slices.Contains(shellGetTools(""), baseTool) {
			return UnknownTool{Tool: baseTool}
		}
		prompt = promptStyle.Render(fmt.Sprintf("flight %s» ", baseTool))
	}
	rlConfig := &readline.Config{
		Prompt: prompt,
		// TODO: Use an XDG cache/data dir for history file.
		HistoryFile:       "/tmp/readline.tmp",
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
