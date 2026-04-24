package main

import (
	"fmt"
	"os/user"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/concertim/flight-user-suite/flight/cliui"
)

func sessionStarted(session *Session) {
	bold := lipgloss.NewStyle().Bold(true)
	var name string
	if session.Name != "" {
		name = lipgloss.JoinHorizontal(lipgloss.Top, "(", bold.Render(session.Name), ") ")
	}
	out := lipgloss.JoinHorizontal(
		lipgloss.Top,
		"A new desktop session ",
		name,
		"has been started. To connect, use the details below:",
	)
	out = lipgloss.Wrap(out, maxTextWidth, " ")
	lipgloss.Println(out)
}

func sessionInfo(session *Session) {
	dt := lipgloss.NewStyle().Foreground(cliui.AlcesBlue).Padding(0, 2, 0, 1)
	dd := lipgloss.NewStyle().Foreground(cliui.LightDark(cliui.Primary, cliui.Cream))
	metadata := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.JoinVertical(
			lipgloss.Right,
			dt.Render("Name"),
			dt.Render("Type"),
			dt.Render("Screen size"),
			dt.Render("State"),
			dt.Render("Started at"),
		),
		lipgloss.JoinVertical(
			lipgloss.Left,
			dd.Render(session.Name),
			dd.Render(session.SessionType),
			dd.Render(session.Geometry),
			dd.Render(string(session.SessionState())),
			dd.Render(session.CreatedAt.Format(time.RFC822)),
		),
	)
	out := lipgloss.JoinVertical(
		lipgloss.Left,
		cliui.Header.Render("Session Details"),
		metadata,
	)
	lipgloss.Println(out)
}

func connectionInfo(session *Session) {
	var username string
	u, err := user.Current()
	if u == nil {
		log.Debug("unable to get current user", "err", err)
		username = "USERNAME"
	} else {
		username = u.Username
	}
	ip := session.IP
	vnc := fmt.Sprintf("vnc://%s:%s@%s:%d", username, session.Password, ip, session.Port())
	ipPort := session.PrimaryConnectionString()
	ipDisplay := fmt.Sprintf("%s:%s", ip, session.Display())

	out := lipgloss.JoinVertical(
		lipgloss.Left,
		cliui.Header.Render("Connection Details"),
		cliui.Subheader.Render("1. Connection address"),
		cliui.Paragraph.Render(lipgloss.Wrap("Copy and paste the first address into your VNC viewer. If that doesn't work, try one of the alternatives:", maxTextWidth, " -")),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.JoinVertical(
				lipgloss.Left,
				cliui.Bullet.Render("* Primary:"),
				cliui.Bullet.Render("* Alternative 1:"),
				cliui.Bullet.Render("* Alternative 2:"),
			),
			lipgloss.JoinVertical(
				lipgloss.Left,
				ipPort,
				vnc,
				ipDisplay,
			),
		),
		cliui.Subheader.PaddingTop(1).Render("2. Password"),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			cliui.Paragraph.Render("If prompted, use this password:"),
			cliui.Code.Render(session.Password),
		),
		cliui.Subheader.Render("3. Need Help?"),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			cliui.Paragraph.MarginBottom(0).Render("For more details on how to connect, run:"),
			cliui.Code.Render("flight howto show flight-desktop"),
		),
	)
	lipgloss.Println(out)
}

func managementInfo(session *Session) {
	instructions := "To view details or stop this session, you will need the Session Name:"
	if session.SessionState() == Exited || session.SessionState() == Broken {
		instructions = "To view details or clean this session, you will need the Session Name:"
	}
	out := lipgloss.JoinVertical(
		lipgloss.Left,
		cliui.Header.Render("Manage this session"),
		cliui.Paragraph.PaddingBottom(0).Render(instructions),
		cliui.Code.Margin(0, 0, 1, 1).Render(session.Name),
		cliui.Paragraph.Render("(Tip: Run 'flight desktop --help' to see management commands)"),
	)
	lipgloss.Println(out)
}
