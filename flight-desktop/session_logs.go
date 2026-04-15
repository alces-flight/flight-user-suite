package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func showSessionLogsCommand() *cli.Command {
	return &cli.Command{
		Name:        "logs",
		Usage:       "Show desktop session logs",
		Description: wordwrap.String("Display logs for a desktop session.", maxTextWidth),
		Category:    "Sessions",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "name", UsageText: "<name>"},
		},
		Before:        assertArgPresent("name"),
		ShellComplete: completeSessionNames,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.StringArg("name")
			session, err := loadSession(name)
			if err != nil {
				return err
			}

			logFile := filepath.Join(session.sessionDir(), "session.log")

			log, err := os.Open(logFile)

			if err != nil {
				return fmt.Errorf("trying to read log file: %w", err)
			}
			defer log.Close()

			scanner := bufio.NewScanner(log)

			for scanner.Scan() {
				line := scanner.Text()
				fmt.Println(line)
			}

			return nil
		},
	}
}
