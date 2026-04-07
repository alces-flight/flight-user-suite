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

func ColourisedHelpPrinter(origHelpPrinter cli.HelpPrinterFunc) cli.HelpPrinterFunc {
	return func(w io.Writer, templ string, data any) {
		var buf bytes.Buffer
		origHelpPrinter(&buf, templ, data)
		bytes, err := io.ReadAll(&buf)
		if err != nil {
			log.Fatal("error formatting help output", "err", err)
		}
		headers := regexp.MustCompile("(?m:^[[:word:]].*:)")
		b := &marker.MarkBuilder{}
		alcesBlue := color.RGB(32, 159, 206)
		out := b.SetString(string(bytes)).
			Mark(marker.MatchRegexp(headers), alcesBlue).
			Build()
		fmt.Fprint(w, out)
	}
}
