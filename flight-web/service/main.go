package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"syscall"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

var (
	// version, commit and date are overwritten at build time.
	version string = "dev"
	commit  string = "unknown"
	date    string = "unknown"

	// Flags
	port         = flag.Int("port", 8080, "port to listen on")
	pidfile      = flag.String("pidfile", "", "pidfile")
	printVersion = flag.Bool("version", false, "print the version")
)

func init() {
	// TODO: Setup log/slog. Save logs to file/stdout?

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

	e := echo.New()
	e.Use(middleware.RequestLogger())

	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	if err := e.Start(address); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}

func writePidfile(path string, pid int) error {
	existingPID, err := readPidfile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if existingPID != 0 {
		return fmt.Errorf("process with PID %d is still running", existingPID)
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0o644)
}

// Read the PID file at path. Return the PID contained in the file if it
// contains a valid PID for a running process.  Otherwise return 0.
func readPidfile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("reading pidfile: %w", err)
	}
	pid, err := strconv.Atoi(string(bytes.TrimSpace(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid content: %w", err)
	}
	if pid == 0 {
		return 0, fmt.Errorf("invalid content: pid=0")
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, nil
	}
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return 0, nil
	}
	return pid, nil
}
