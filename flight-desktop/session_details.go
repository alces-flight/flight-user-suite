package main

import (
	"fmt"
	"os/user"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/concertim/flight-user-suite/flight/pkg"
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
	dt := lipgloss.NewStyle().Foreground(pkg.AlcesBlue).Padding(0, 2, 0, 1)
	dd := lipgloss.NewStyle().Foreground(pkg.LightDark(pkg.Primary, pkg.Cream))
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
		pkg.Header.Render("Session Details"),
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
		pkg.Header.Render("Connection Details"),
		pkg.Subheader.Render("1. Connection address"),
		pkg.Paragraph.Render(lipgloss.Wrap("Copy and paste the first address into your VNC viewer. If that doesn't work, try one of the alternatives:", maxTextWidth, " -")),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.JoinVertical(
				lipgloss.Left,
				pkg.Bullet.Render("* Primary:"),
				pkg.Bullet.Render("* Alternative 1:"),
				pkg.Bullet.Render("* Alternative 2:"),
			),
			lipgloss.JoinVertical(
				lipgloss.Left,
				ipPort,
				vnc,
				ipDisplay,
			),
		),
		pkg.Subheader.PaddingTop(1).Render("2. Password"),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			pkg.Paragraph.Render("If prompted, use this password:"),
			pkg.Code.Render(session.Password),
		),
		pkg.Subheader.Render("3. Need Help?"),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			pkg.Paragraph.MarginBottom(0).Render("For more details on how to connect, run:"),
			pkg.Code.Render("flight howto show flight-desktop"),
		),
	)
	lipgloss.Println(out)
}

func managementInfo(session *Session) {
	out := lipgloss.JoinVertical(
		lipgloss.Left,
		pkg.Header.Render("Manage this session"),
		pkg.Paragraph.PaddingBottom(0).Render("To view details or stop this session, you will need the Session Name:"),
		pkg.Code.Margin(0, 0, 1, 1).Render(session.Name),
		pkg.Paragraph.Render("(Tip: Run 'flight desktop --help' to see management commands)"),
	)
	lipgloss.Println(out)
}
