package main

import (
	"context"
	"fmt"

	"charm.land/log/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
	"github.com/yarlson/pin"
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
			p := pin.New(fmt.Sprintf("Killing desktop session %s", session.ID),
				pin.WithSpinnerColor(pin.ColorCyan),
				pin.WithTextColor(pin.ColorGreen),
				pin.WithDoneSymbol('\u2705'),
				pin.WithFailSymbol('\u274c'),
				pin.WithFailColor(pin.ColorRed),
			)
			cancel := p.Start(ctx)
			defer cancel()
			err = session.Kill(ctx)
			if err != nil {
				p.Fail("Terminating session failed")
				return fmt.Errorf("terminating session: %w", err)
			}
			p.Stop("Session terminated")
			fmt.Println()
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
