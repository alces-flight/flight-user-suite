package main

import (
	"context"
	"fmt"
	"time"

	"charm.land/log/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func webifySessionCommand() *cli.Command {
	return &cli.Command{
		Name:        "webify",
		Usage:       "Start web access support for an active desktop session",
		Description: wordwrap.String("Start required and optional web support programs for an active interactive desktop session. Can be used to add web support to sessions that were started before web dependencies were installed.", maxTextWidth),
		Category:    "Sessions",
		// Flags:       []cli.Flag{formatFlag},
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "name", UsageText: "<name>"},
		},
		Before:        assertArgPresent("name"),
		ShellComplete: completeActiveSessionNames,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.StringArg("name")
			session, err := loadSession(name)
			// if cmd.String("format") == "json" {
			// 	return writeSessionJSON(session, err)
			// }
			if err != nil {
				if err2 := session.RemoveSessionDir(); err2 != nil {
					log.Debug("Removing session dir", "sessionDir", session.sessionDir(), "err", err2)
				}
				return err
			}
			switch session.ComputedState() {
			case "active":
				p := createPin("Starting web access support...")
				cancel := p.Start(ctx)
				defer cancel()
				timer := time.After(1 * time.Second)
				err = session.Webify(ctx)
				<-timer
				if err != nil {
					p.Fail("Starting web access support failed")
					return err
				}
				p.Stop("Web access support is ready!")
				return nil

			case "remote":
				return fmt.Errorf("session %s is not local", session.Name)
			default:
				return fmt.Errorf("session %s is not active", session.Name)
			}
		},
	}
}
