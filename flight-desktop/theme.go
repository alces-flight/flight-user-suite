package main

import (
	"os"

	"charm.land/lipgloss/v2"
)

var (
	black     = lipgloss.Color("#000000")
	cream     = lipgloss.Color("#FFFDF5")
	ctmOrange = lipgloss.Color("#ff7401")
	grey      = lipgloss.Color("#CCD0DA")
	primary   = lipgloss.Color("#2C3E50")

	hasDark   = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark = lipgloss.LightDark(hasDark)

	tableHeaderStyle  = lipgloss.NewStyle().Foreground(ctmOrange).Bold(true).Align(lipgloss.Center)
	tableCellStyle    = lipgloss.NewStyle().Padding(0, 1)
	tableOddRowStyle  = tableCellStyle.Foreground(lightDark(black, grey))
	tableEvenRowStyle = tableCellStyle.Foreground(lightDark(primary, cream))
)
