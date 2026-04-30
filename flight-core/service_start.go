package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
	"github.com/yarlson/pin"
)

func startCommand() *cli.Command {
	return &cli.Command{
		Name:        "start",
		Usage:       "Start the Flight Web Suite service",
		Description: wordwrap.String("Starts the Flight Web Suite service, if it is not already running.", maxTextWidth),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Exec the service, print where it is running and exit.
			service := Service{ID: "web-suite", Name: "Web Suite"}

			p := pin.New(
				fmt.Sprintf("Starting service %s...", service.Name),
				pin.WithSpinnerColor(pin.ColorCyan),
				pin.WithTextColor(pin.ColorGreen),
				pin.WithDoneSymbol('\u2705'),
				pin.WithFailSymbol('\u274c'),
				pin.WithFailColor(pin.ColorRed),
			)
			cancel := p.Start(ctx)
			defer cancel()

			// Pause for better spinner UX
			<-time.After(1 * time.Second)

			err := service.Start(ctx)
			if err != nil {
				p.Fail(fmt.Sprintf("Starting %s service failed: %s", service.Name, err))
				os.Exit(1)
			}
			// TODO: Wait for process to have started and reported the address
			// its listening on.
			// * Could create a file descriptor and have process report on that.
			// * Could have the process save its URL to a file.
			address := fmt.Sprintf("0.0.0.0:%d", 8080)
			p.Stop(fmt.Sprintf("Started %s. Listening on %s\n", service.Name, address))
			return nil
		},
	}
}
