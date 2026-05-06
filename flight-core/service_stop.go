package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/concertim/flight-user-suite/flight/userroles"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
	"github.com/yarlson/pin"
)

func stopCommand() *cli.Command {
	return &cli.Command{
		Name:        "stop",
		Usage:       "Stop the Flight Web Suite service",
		Description: wordwrap.String("Stops the Flight Web Suite service, if it is running.", maxTextWidth),
		Action:      userroles.AsRoot(stopService),
	}
}

func stopService(ctx context.Context, cmd *cli.Command) error {
	service := Service{ID: "web-suite", Name: "Web Suite"}
	p := pin.New(
		fmt.Sprintf("Stopping %s service...", service.Name),
		pin.WithSpinnerColor(pin.ColorCyan),
		pin.WithTextColor(pin.ColorGreen),
		pin.WithDoneSymbol('\u2705'),
		pin.WithFailSymbol('\u274c'),
		pin.WithFailColor(pin.ColorRed),
	)
	cancel := p.Start(ctx)
	defer cancel()

	timer := time.After(1 * time.Second)

	err := service.Kill()
	// Pause for better spinner UX
	<-timer
	if err != nil {
		p.Fail(fmt.Sprintf("Stopping %s service failed: %s", service.Name, err))
		os.Exit(1)
	}
	p.Stop(fmt.Sprintf("%s service stopped\n", service.Name))
	return nil
}
