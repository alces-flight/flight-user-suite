package main

import (
	"context"
	"fmt"

	"charm.land/log/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func killSessionCommand() *cli.Command {
	return &cli.Command{
		Name:        "kill",
		Usage:       "Terminate an interactive desktop session",
		Description: wordwrap.String(fmt.Sprintf("Instruct an active interactive desktop session to terminate.\n\nThe <id> parameter should specify the session identity, use '%s list' to see a list of your sessions.", progName), 80),
		Category:    "Sessions",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "id", UsageText: "<id>"},
		},
		Before: assertArgPresent("id"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			id := cmd.StringArg("id")
			session, err := loadSession(id)
			if err != nil {
				if err2 := session.RemoveSessionDir(); err2 != nil {
					log.Debug("Removing session dir", "sessionDir", session.sessionDir(), "err", err2)
				}
				return err
			}
			// TODO: Display a spinner.
			fmt.Printf("Killing desktop session %s\n", session.ID)
			err = session.Kill(ctx)
			// TODO: Stop the spinner
			if err != nil {
				fmt.Printf("\u274c Terminating session\n\n")
				return fmt.Errorf("terminating session: %w", err)
			}
			fmt.Printf("\u2705 Terminating session\n\n")
			fmt.Printf("Desktop session '%s' has been terminated.\n", session.ID)
			return nil
		},
	}
}

type UnknownSession struct {
	Session string
}

func (us UnknownSession) Error() string {
	return fmt.Sprintf("Unknown session: %s", us.Session)
}
