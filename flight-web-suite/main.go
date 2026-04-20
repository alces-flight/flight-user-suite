package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

var (
	// version, commit and date are overwritten at build time.
	version string = "dev"
	commit  string = "unknown"
	date    string = "unknown"

	flightRoot string = "/opt/flight"

	// Flags
	port         = flag.Int("port", 8080, "port to listen on")
	pidfile      = flag.String("pidfile", "", "pidfile")
	printVersion = flag.Bool("version", false, "print the version")
)

type Tool struct {
	Name        string
	Description string
	URL         string
	IconPath    string
}

func init() {
	// TODO: Setup log/slog. Save logs to file/stdout?

	if root, ok := os.LookupEnv("FLIGHT_ROOT"); ok {
		flightRoot = root
	}

	flag.Usage = func() {
		cmd := path.Base(os.Args[0])
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage: %s [OPTION]...\n", cmd) // nolint:errcheck
		fmt.Fprintf(w, "\nOPTIONS:\n")                 // nolint:errcheck
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if *printVersion {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "version=%s revision=%s date=%s\n", version, commit, date) // nolint:errcheck
		os.Exit(0)
	} else if len(flag.Args()) > 0 {
		flag.Usage()
		os.Exit(1)
	}

	if *pidfile != "" {
		err := writePidfile(*pidfile, os.Getpid())
		if err != nil {
			w := flag.CommandLine.Output()
			fmt.Fprintf(w, "Unable to write pidfile: %s", err.Error()) // nolint:errcheck
			os.Exit(1)
		}
	}

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	log.Printf("Starting Web Suite on %s\n", address)

	e := newApp()
	if err := e.Start(address); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}

func newApp() *echo.Echo {
	e := echo.New()
	e.Pre(MethodOverrideMiddleware())
	e.Use(middleware.RequestLogger())
	e.Use(NewSessionMiddleware())

	// TODO: Standardise on extension for Go html/template files.
	t := template.Must(template.ParseGlob(getDirectory("views") + "/*.html"))
	// t = template.Must(t.ParseGlob(getDirectory("views") + "/*/*.html"))
	// t = template.Must(t.ParseGlob(getDirectory("views") + "/*.gohtml"))
	t = template.Must(t.ParseGlob(getDirectory("views") + "/*/*.gohtml"))
	e.Renderer = &echo.TemplateRenderer{Template: t}

	e.Static("/assets", getDirectory("assets"))
	e.Static("/static", getDirectory("static"))
	e.GET("/", func(c *echo.Context) error {
		data := indexData()
		return c.Render(http.StatusOK, "home", AddCommonData(c, data))
	})
	e.GET("/sessions", newSessionHandler)
	e.POST("/sessions", createSessionHandler)
	e.DELETE("/sessions", destroySessionHandler)

	e.GET("/websockify", func(c *echo.Context) error { return nil }, NewWSProxyMiddleware())

	return e
}

func indexData() map[string]any {
	return map[string]any{
		"EnvName": "My Cluster",
		"Tools": []Tool{
			{
				Name:        "Flight Desktop",
				Description: "Access interactive desktop sessions",
				URL:         "/desktop",
				IconPath:    "/assets/images/desktop.png",
			},
			{
				Name:        "Flight Howto",
				Description: "Learn about the Flight User Suite and using your cluster",
				URL:         "/howto",
				IconPath:    "/assets/images/howto.png",
			},
		},
	}
}

func getDirectory(dirName string) string {
	// If the named directory exists in our CWD, use that; otherwise use the
	// expected locations in a deployed Flight User Suite installation.
	// Net effect is running from "local" files in development, and as expected
	// in production.
	_, err := os.Stat(dirName)
	if errors.Is(err, os.ErrNotExist) {
		return path.Join(flightRoot, "var", "web-suite", dirName)
	}
	return dirName
}
