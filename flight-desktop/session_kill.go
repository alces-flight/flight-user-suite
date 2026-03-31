package main

import (
	"context"
	"fmt"
	"os"

	"charm.land/log/v2"
	"github.com/google/uuid"
	"github.com/muesli/reflow/wordwrap"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

func killSessionCommand() *cli.Command {
	return &cli.Command{
		Name:        "kill",
		Usage:       "Terminate an interactive desktop session",
		Description: wordwrap.String(fmt.Sprintf("Instruct an active interactive desktop session to terminate.\n\nThe <id> parameter should specify the session identity, use '%s list' to see a list of your sessions.", progName), 80),
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
			fmt.Printf("Killing desktop session %s\n", session.UUID.String())
			err = session.Kill(ctx)
			// TODO: Stop the spinner
			if err != nil {
				fmt.Printf("\u274c Terminating session\n\n")
				return fmt.Errorf("terminating session: %w", err)
			}
			fmt.Printf("\u2705 Terminating session\n\n")
			fmt.Printf("Desktop session '%s' has been terminated.\n", session.UUID.String())
			return nil
		},
	}
}

func loadSession(id string) (*Session, error) {
	session := &Session{UUID: uuid.MustParse(id)}
	log.Debug("Loading session", "sessionDir", session.sessionDir())
	info, err := os.Stat(session.sessionDir())
	if err != nil {
		log.Debug("Error checking session dir", "sessionDir", session.sessionDir(), "err", err)
		session.SessionState = Broken
		return session, UnknownSession{Session: id}
	}
	if !info.IsDir() {
		log.Debug("Session dir is not a directory", "sessionDir", session.sessionDir())
		session.SessionState = Broken
		return session, UnknownSession{Session: id}
	}

	data, err := os.ReadFile(session.metadataFile())
	if err != nil {
		log.Debug("Reading session metadata", "metadataFile", session.metadataFile(), "err", err)
		session.SessionState = Broken
		return session, nil
	}
	err = yaml.Unmarshal(data, &session)
	if err != nil {
		log.Debug("Loading session metadata", "metadataFile", session.metadataFile(), "err", err)
		session.SessionState = Broken
		return session, nil
	}
	return session, nil
}

type UnknownSession struct {
	Session string
}

func (us UnknownSession) Error() string {
	return fmt.Sprintf("Unknown session: %s", us.Session)
}
