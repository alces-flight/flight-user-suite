package pkg

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"

	"github.com/cyucelen/marker"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

func OrangifiedHelpPrinter() func(w io.Writer, templ string, data any) {
	origHelpPrinter := cli.HelpPrinter
	return func(w io.Writer, templ string, data any) {
		var buf bytes.Buffer
		origHelpPrinter(&buf, templ, data)
		bytes, err := io.ReadAll(&buf)
		if err != nil {
			log.Fatal("error formatting help output", "err", err)
		}
		headers := regexp.MustCompile("(?m:^[[:word:]].*:)")
		b := &marker.MarkBuilder{}
		ctmOrange := color.RGB(255, 116, 1)
		out := b.SetString(string(bytes)).
			Mark(marker.MatchRegexp(headers), ctmOrange).
			Build()
		fmt.Fprint(w, out)
	}
}
