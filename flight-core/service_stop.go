package main

import (
	"context"
	"fmt"

	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
)

func stopCommand() *cli.Command {
	return &cli.Command{
		Name:        "stop",
		Usage:       "Stop the Flight Web Suite service",
		Description: wordwrap.String("Stops the Flight Web Suite service, if it is running.", maxTextWidth),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			service := Service{ID: "web-suite", Name: "Web Suite"}
			fmt.Printf("Stopping %s service...\n", service.Name)
			err := service.Kill()
			if err != nil {
				return cli.Exit(fmt.Sprintf("Error stopping service: %s", err.Error()), 1)
			}
			fmt.Printf("%s service stopped\n", service.Name)
			return nil
		},
	}
}
