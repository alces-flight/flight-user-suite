package main

import (
	"fmt"
	"os/user"
	"time"

	"charm.land/log/v2"
	"github.com/muesli/reflow/wordwrap"
)

func printSessionStartedMessage(session Session) {
	var username string
	u, err := user.Current()
	if u == nil {
		log.Debug("unable to get current user", "err", err)
		username = "USERNAME"
	} else {
		username = u.Username
	}
	ip := session.PrimaryIP().String()
	vnc := fmt.Sprintf("vnc://%s:%s@%s:%d", username, session.Password, ip, session.Port())
	ipPort := session.PrimaryConnectionString()
	ipDisplay := fmt.Sprintf("%s:%s", ip, session.Display())

	var name string
	if session.Name != "" {
		name = fmt.Sprintf("(%s) ", session.Name)
	}
	out := fmt.Sprintf("A new destop session %shas been started. To connect, use the details below:\n\n1. Connection Address\n\nCopy and paste the first address into your VNC viewer. If that doesn't work, try one of the alternatives:\n\n* Primary:       %s\n* Alternative 1: %s\n* Alternative 2: %s\n\n2. Password\n\nIf prompted, use this password: %s\n\n3. Need Help?\n\nFor a more details on how to connect, run: 'flight howto show flight-desktop'\n\n4. Manage this session\n\nTo view details or stop this session later, you will need the Session ID:\n\n    %s\n\n(Tip: Run 'flight desktop --help' to see management commands)\n\nStarted at: %s\n", name, ipPort, vnc, ipDisplay, session.Password, session.ID, session.CreatedAt.Format(time.RFC822))
	out = wordwrap.String(out, 80)
	fmt.Print(out)
}
