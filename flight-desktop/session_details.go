package main

import (
	"fmt"
	"os/user"
	"time"

	"charm.land/log/v2"
	"github.com/muesli/reflow/wordwrap"
)

func printSessionDetails(session Session) {
	// TODO: Better output for TTY.
	fmt.Println()
	fmt.Printf("Identity\t%s\n", session.UUID)
	fmt.Printf("Name\t\t%s\n", session.Name)
	fmt.Printf("Type\t\t%s\n", session.SessionType)
	fmt.Printf("Host IP\t\t%s\n", session.PrimaryIP())
	fmt.Printf("Hostname\t%s\n", session.Metadata.Host)
	fmt.Printf("Port\t\t%d\n", session.Metadata.Port())
	fmt.Printf("Display\t\t:%s\n", session.Metadata.Display)
	fmt.Printf("Password\t%s\n", session.Password)
	fmt.Printf("State\t\t%s\n", session.SessionState)
	fmt.Printf("Created at\t%s\n", session.CreatedAt.Format(time.RFC822))
	fmt.Printf("Geometry\t%s\n", session.Geometry)
	fmt.Println()
}

func accessSummary(s Session) {
	isPublic := !s.PrimaryIP().IsPrivate() && portIsReachable(s.Metadata.Port())

	var prefix string
	var suffix string

	if isPublic {
		prefix = "This desktop session is directly accessible from the public internet."
		suffix = wordwrap.String("Accessing desktop  sessions directly is NOT SECURE and we highly recommend using a secure port forwarding technique with 'ssh' to secure your desktop session.", 80)
	} else {
		prefix = wordwrap.String("This desktop session is not accessible from the public internet, but may be directly accessible from within your local network or over a virtual private network (VPN).", 80)
		suffix = ""
	}

	var username string
	u, err := user.Current()
	if err != nil {
		log.Debug("unable to get current user", "err", err)
		username = "USERNAME"
	}
	username = u.Username
	ip := s.PrimaryIP().String()
	vnc := fmt.Sprintf("\tvnc://%s:%s@%s:%d", username, s.Password, ip, s.Metadata.Port())
	ipPort := fmt.Sprintf("\t%s:%d", ip, s.Metadata.Port())
	ipDisplay := fmt.Sprintf("\t%s:%s", ip, s.Metadata.Display)

	details := wordwrap.String("Depending on your client and network configuration you may be able to directly connect to the session using:", 80)
	details = fmt.Sprintf("%s\n\n%s\n%s\n%s", details, vnc, ipPort, ipDisplay)

	fmt.Printf("%s\n\n", prefix)
	fmt.Printf("%s\n\n", details)
	if suffix != "" {
		fmt.Printf("%s\n\n", suffix)
	}
	fmt.Printf("If prompted, you should supply the following password: %s\n", s.Password)
}
