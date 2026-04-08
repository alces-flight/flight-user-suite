package pkg

import (
	"os"

	"charm.land/lipgloss/v2"
)

var (
	hasDark   = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	LightDark = lipgloss.LightDark(hasDark)

	Black     = lipgloss.Color("#000000")
	Cream     = lipgloss.Color("#FFFDF5")
	AlcesBlue = lipgloss.Color("#209FCE")
	LightBlue = lipgloss.Color("#83D0ED")
	Grey      = lipgloss.Color("#CCD0DA")
	Primary   = lipgloss.Color("#2C3E50")

	Header    = lipgloss.NewStyle().Margin(1, 1, 1, 0).Bold(true).Foreground(AlcesBlue)
	Subheader = lipgloss.NewStyle().Margin(0, 1, 1, 1).Bold(true)
	Paragraph = lipgloss.NewStyle().Margin(0, 1, 1, 1)
	Code      = lipgloss.NewStyle().Background(Grey).Foreground(Primary).Bold(true)
	Bullet    = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(2).Bold(true)
	Hyperlink = lipgloss.NewStyle().Foreground(LightBlue)

	TableHeaderStyle  = lipgloss.NewStyle().Foreground(LightBlue).Bold(true).Align(lipgloss.Center)
	TableCellStyle    = lipgloss.NewStyle().Padding(0, 1)
	TableOddRowStyle  = TableCellStyle.Foreground(LightDark(Black, Grey))
	TableEvenRowStyle = TableCellStyle.Foreground(LightDark(Primary, Cream))
)
