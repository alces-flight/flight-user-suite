package main

import (
	"context"

	"charm.land/log/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func showSessionCommand() *cli.Command {
	return &cli.Command{
		Name:        "show",
		Usage:       "Show information about a desktop session",
		Description: wordwrap.String("Display the connection information for a desktop session.", maxTextWidth),
		Category:    "Sessions",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "name", UsageText: "<name>"},
		},
		Before: assertArgPresent("name"),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.StringArg("name")
			session, err := loadSession(name)
			if err != nil {
				if err2 := session.RemoveSessionDir(); err2 != nil {
					log.Debug("Removing session dir", "sessionDir", session.sessionDir(), "err", err2)
				}
				return err
			}
			sessionInfo(session)
			if session.State != Exited && session.State != Broken {
				connectionInfo(session)
			}
			managementInfo(session)
			return nil
		},
	}
}
