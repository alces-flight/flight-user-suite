package main

import (
	"fmt"
	"os/user"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
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
	dt := lipgloss.NewStyle().Foreground(ctmOrange).Padding(0, 2, 0, 1)
	dd := lipgloss.NewStyle().Foreground(lightDark(primary, cream))
	metadata := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.JoinVertical(
			lipgloss.Right,
			dt.Render("Name"),
			dt.Render("Type"),
			dt.Render("Screen size"),
			dt.Render("Started at"),
		),
		lipgloss.JoinVertical(
			lipgloss.Left,
			dd.Render(session.Name),
			dd.Render(session.SessionType),
			dd.Render(session.Geometry),
			dd.Render(session.CreatedAt.Format(time.RFC822)),
		),
	)
	out := lipgloss.JoinVertical(
		lipgloss.Left,
		header.Render("Session Details"),
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
	ip := session.PrimaryIP().String()
	vnc := fmt.Sprintf("vnc://%s:%s@%s:%d", username, session.Password, ip, session.Port())
	ipPort := session.PrimaryConnectionString()
	ipDisplay := fmt.Sprintf("%s:%s", ip, session.Display())

	out := lipgloss.JoinVertical(
		lipgloss.Left,
		header.Render("Connection Details"),
		subheader.Render("1. Connection address"),
		paragraph.Render(lipgloss.Wrap("Copy and paste the first address into your VNC viewer. If that doesn't work, try one of the alternatives:", maxTextWidth, " -")),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.JoinVertical(
				lipgloss.Left,
				bullet.Render("* Primary:"),
				bullet.Render("* Alternative 1:"),
				bullet.Render("* Alternative 2:"),
			),
			lipgloss.JoinVertical(
				lipgloss.Left,
				ipPort,
				vnc,
				ipDisplay,
			),
		),
		subheader.PaddingTop(1).Render("2. Password"),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			paragraph.Render("If prompted, use this password:"),
			code.Render(session.Password),
		),
		subheader.Render("3. Need Help?"),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			paragraph.Render("For more details on how to connect, run:"),
			code.Render("flight howto show flight-desktop"),
		),
		header.MarginTop(0).Render("Manage this session"),
		paragraph.PaddingBottom(0).Render("To view details or stop this session later, you will need the Session ID:"),
		code.Margin(0, 0, 1, 1).Render(session.ID),
		paragraph.Render("(Tip: Run 'flight desktop --help' to see management commands)"),
	)
	lipgloss.Println(out)
}
