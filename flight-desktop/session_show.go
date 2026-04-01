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
			sessionInfo(session)
			if session.SessionState == Active {
				connectionInfo(session)
			}
			managementInfo(session)
			return nil
		},
	}
}
