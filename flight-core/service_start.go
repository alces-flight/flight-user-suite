package main

import (
	"context"
	"fmt"

	"charm.land/log/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func startCommand() *cli.Command {
	return &cli.Command{
		Name:        "start",
		Usage:       "Start the Flight Web Suite service",
		Description: wordwrap.String("Starts the Flight Web Suite service, if it is not already running.", maxTextWidth),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Exec the service, print where it is running and exit.
			service := Service{ID: "web-suite", Name: "Web Suite"}
			fmt.Printf("Starting service %s...\n", service.Name)
			err := service.Start(ctx)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Starting %s service failed: %s", service.Name, err), 1)
			}
			// TODO: Wait for process to have started and reported the address
			// its listening on.
			// * Could create a file descriptor and have process report on that.
			// * Could have the process save its URL to a file.
			address := fmt.Sprintf("0.0.0.0:%d", 8080)
			log.Printf("Started %s. Listening on %s\n", service.Name, address)
			return nil
		},
	}
}
