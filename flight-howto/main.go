package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

// Wrapper around exec.ExitError that avoids the default handling by urfave/cli.
// TODO deduplicate from flight-core/tools.go (both type and func below)
type SilentExitError struct {
	ExitCode  int
	exitError error
}

func (ee SilentExitError) Error() string {
	return ee.exitError.Error()
}

var (
	flightRoot string = "/opt/flight"
	howtoDir   string
)

func init() {
	if root, ok := os.LookupEnv("FLIGHT_ROOT"); ok {
		flightRoot = root
	}
	howtoDir = filepath.Join(flightRoot, "usr", "share", "doc", "howtos-enabled")
}

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list available howtos",
				Action:  list,
			},
			{
				Name:      "show",
				Aliases:   []string{"s"},
				Usage:     "show a howto",
				ArgsUsage: "<howto>",
				Action:    show,
			},
		},
	}

	// TODO deduplicate this from equivalent section in flight-core/main.go?
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		// A bunch of checks to avoid reporting the usage errors twice.
		errStr := err.Error()
		if (strings.Contains(errStr, "invalid value") && strings.Contains(errStr, "for flag")) ||
			(strings.Contains(errStr, "flag provided but not defined")) ||
			(strings.Contains(errStr, "flag needs an argument")) {
			// We've already reported the usage error.  No need to do so a
			// second time.
			os.Exit(1)
		}

		if strings.Contains(errStr, "cannot be set along with") {
			log.Printf("\nIncorrect Usage: %s", err)
			os.Exit(1)
		}

		if exitError, ok := errors.AsType[SilentExitError](err); ok {
			os.Exit(exitError.ExitCode)
		} else {
			log.Printf("%s\n", err)
			os.Exit(1)
		}
	}
}

func list(ctx context.Context, cmd *cli.Command) error {
	return _print_dir_contents(howtoDir)
}

func show(ctx context.Context, cmd *cli.Command) error {
	fullPath := filepath.Join(howtoDir, cmd.Args().First())
	markdown, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}
	fmt.Println(string(markdown))
	return nil
}

func _print_dir_contents(dir_path string) error {
	file, err := os.Open(dir_path)
	if err != nil {
		return err
	}
	defer file.Close()
	names, _ := file.Readdirnames(0)
	for _, name := range names {
		filePath := fmt.Sprintf("%v/%v", dir_path, name)
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		fileInfo, err := file.Stat()
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(howtoDir, filePath)
		if err != nil {
			return err
		}
		ext := filepath.Ext(relPath)
		if ext == ".md" {
			fmt.Println(relPath)
		}
		if fileInfo.IsDir() {
			_print_dir_contents(filePath)
		}
	}
	return nil
}
