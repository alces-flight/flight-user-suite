module github.com/concertim/flight-user-suite/flight-desktop

go 1.26.1

require (
	charm.land/log/v2 v2.0.0
	github.com/concertim/flight-user-suite/flight v0.0.0-00010101000000-000000000000
	github.com/muesli/reflow v0.3.0
	github.com/urfave/cli/v3 v3.8.0
	golang.org/x/term v0.41.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/cyucelen/marker v0.0.0-20220628090808-ec8d542c2d28 // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/yarlson/pin v0.9.1
)

require (
	charm.land/lipgloss/v2 v2.0.1
	github.com/adrg/xdg v0.5.3
	github.com/charmbracelet/colorprofile v0.4.2 // indirect
	github.com/charmbracelet/ultraviolet v0.0.0-20251205161215-1948445e3318 // indirect
	github.com/charmbracelet/x/ansi v0.11.6 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/charmbracelet/x/termios v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/go-logfmt/logfmt v0.6.1 // indirect
	github.com/google/uuid v1.6.0
	github.com/lucasb-eyer/go-colorful v1.3.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
)

replace github.com/concertim/flight-user-suite/flight => ../flight-core
