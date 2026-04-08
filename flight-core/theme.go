package main

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	fatihColor "github.com/fatih/color"
)

var (
	hasDark   = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark = lipgloss.LightDark(hasDark)

	black     = lipgloss.Color("#000000")
	cream     = lipgloss.Color("#FFFDF5")
	alcesBlue = lipgloss.Color("#209FCE")
	grey      = lipgloss.Color("#CCD0DA")
	primary   = lipgloss.Color("#2C3E50")

	promptStyle = lipgloss.NewStyle().Foreground(alcesBlue)

	tableHeaderStyle  = lipgloss.NewStyle().Foreground(alcesBlue).Bold(true).Align(lipgloss.Center)
	tableCellStyle    = lipgloss.NewStyle().Padding(0, 1)
	tableOddRowStyle  = tableCellStyle.Foreground(lightDark(black, grey))
	tableEvenRowStyle = tableCellStyle.Foreground(lightDark(primary, cream))
)

// Best effort attempt to convert from lipgloss.Color to fatish/color.Color.
// This will not work correctly if the lipgloss color has an alpha value set.
func imageColorToFatihColor(src color.Color) *fatihColor.Color {
	r, g, b, a := src.RGBA()
	if a != 65535 {
		log.Warn("color conversion with alpha value not supported", "src", src)
	}
	r = r >> 8
	g = g >> 8
	b = b >> 8
	return fatihColor.RGB(int(r), int(g), int(b))
}
