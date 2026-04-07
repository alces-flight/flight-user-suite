package main

import (
	"os"

	"charm.land/lipgloss/v2"
)

var (
	hasDark   = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark = lipgloss.LightDark(hasDark)

	black     = lipgloss.Color("#000000")
	cream     = lipgloss.Color("#FFFDF5")
	alcesBlue = lipgloss.Color("#209FCE")
	grey      = lipgloss.Color("#CCD0DA")
	primary   = lipgloss.Color("#2C3E50")

	header    = lipgloss.NewStyle().Margin(1, 1, 1, 0).Bold(true).Foreground(alcesBlue)
	subheader = lipgloss.NewStyle().Margin(0, 1, 1, 1).Bold(true)
	paragraph = lipgloss.NewStyle().Margin(0, 1, 1, 1)
	code      = lipgloss.NewStyle().Background(grey).Foreground(primary).Bold(true)
	bullet    = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(2).Bold(true)
	hyperlink = lipgloss.NewStyle().Foreground(lipgloss.Cyan)

	tableHeaderStyle  = lipgloss.NewStyle().Foreground(alcesBlue).Bold(true).Align(lipgloss.Center)
	tableCellStyle    = lipgloss.NewStyle().Padding(0, 1)
	tableOddRowStyle  = tableCellStyle.Foreground(lightDark(black, grey))
	tableEvenRowStyle = tableCellStyle.Foreground(lightDark(primary, cream))
)
