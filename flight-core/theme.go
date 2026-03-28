package main

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	fatihColor "github.com/fatih/color"
)

var (
	ctmOrange   = lipgloss.Color("#ff7401")
	promptStyle = lipgloss.NewStyle().Foreground(ctmOrange)
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
