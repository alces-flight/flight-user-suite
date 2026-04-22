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
	"path/filepath"

	"github.com/concertim/flight-user-suite/flight/configenv"
	"github.com/concertim/flight-user-suite/flight/pidfile"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

var (
	// version, commit and date are overwritten at build time.
	version string = "dev"
	commit  string = "unknown"
	date    string = "unknown"

	env               configenv.Env
	authenticatorPath string
	config            webSuiteConfig

	// Flags
	pidfilePath  = flag.String("pidfile", "", "pidfile")
	printVersion = flag.Bool("version", false, "print the version")
)

func init() {
	// TODO: Setup log/slog. Save logs to file/stdout?

	var err error
	env, err = configenv.InitFlightEnv()
	if err != nil {
		panic(fmt.Errorf("initializing flight env: %w", err))
	}
	authenticatorPath = filepath.Join(env.FlightRoot, "usr", "libexec", "web-suite", "authenticate.py")

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

	var err error
	config, err = loadConfig()
	if err != nil {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Unable to load config: %s\n", err) // nolint:errcheck
		os.Exit(1)
	}

	if *pidfilePath != "" {
		err = pidfile.Write(*pidfilePath, os.Getpid())
		if err != nil {
			w := flag.CommandLine.Output()
			fmt.Fprintf(w, "Unable to write pidfile: %s", err.Error()) // nolint:errcheck
			os.Exit(1)
		}
	}

	address := fmt.Sprintf("0.0.0.0:%d", config.Port)
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

	t := template.Must(template.ParseGlob(getDirectory("views") + "/*.gohtml"))
	t = template.Must(t.ParseGlob(getDirectory("views") + "/*/*.gohtml"))
	e.Renderer = &echo.TemplateRenderer{Template: t}

	e.Static("/assets", getDirectory("assets"))
	e.Static("/static", getDirectory("static"))
	e.GET("/", func(c *echo.Context) error {
		data, err := indexData()
		if err != nil {
			e.Logger.Error("Error when calling indexData", "error", err)
		}
		return c.Render(http.StatusOK, "home", AddCommonData(c, data))
	})
	e.GET("/desktop", indexDesktopSessionsHandler)
	e.DELETE("/desktop/:sessionName", destroyDesktopSessionHandler)
	e.GET("/sessions", newSessionHandler)
	e.POST("/sessions", createSessionHandler)
	e.DELETE("/sessions", destroySessionHandler)
	return e
}

func indexData() (map[string]any, error) {
	toolsList, err := getToolsList()
	return map[string]any{
		"EnvName":  "My Cluster",
		"Tools":    toolsList,
		"HasTools": len(toolsList) > 0,
	}, err
}

func getToolsList() ([]Tool, error) {
	flightTools, err := getTools(true)

	availableTools := make([]Tool, 0, len(flightTools))
	for _, flightTool := range flightTools {
		toolDef, exists := toolDefs[flightTool.Name]
		if exists {
			availableTools = append(availableTools, toolDef)
		}
	}
	return availableTools, err
}

func getDirectory(dirName string) string {
	// If the named directory exists in our CWD, use that; otherwise use the
	// expected locations in a deployed Flight User Suite installation.
	// Net effect is running from "local" files in development, and as expected
	// in production.
	_, err := os.Stat(dirName)
	if errors.Is(err, os.ErrNotExist) {
		return path.Join(env.FlightRoot, "var", "web-suite", dirName)
	}
	return dirName
}
