package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestKillSessionJSON(t *testing.T) {
	tests := []struct {
		name             string
		sessionName      string
		setup            func(t *testing.T)
		expectedResponse map[string]any
		expectedExitCode int
	}{
		{
			name:        "success",
			sessionName: "alpha",
			setup: func(t *testing.T) {
				proc := startTestProcess(t)
				createDesktopSessionFixture(t, desktopSessionFixture{
					Name:    "alpha",
					IP:      "127.0.0.1",
					Pid:     proc.Process.Pid,
					State:   "active",
					Host:    "localhost",
					Display: "1",
				})
			},
			expectedResponse: map[string]any{
				"success":      true,
				"session_name": "alpha",
			},
			expectedExitCode: 0,
		},
		{
			name:        "not local",
			sessionName: "beta",
			setup: func(t *testing.T) {
				proc := startTestProcess(t)
				createDesktopSessionFixture(t, desktopSessionFixture{
					Name:    "beta",
					IP:      "192.0.2.10",
					Pid:     proc.Process.Pid,
					State:   "active",
					Host:    "remote-host",
					Display: "2",
				})
			},
			expectedResponse: map[string]any{
				"success":      false,
				"session_name": "beta",
				"error":        "Desktop session 'beta' is not local.",
				"reason":       "not_local",
			},
			expectedExitCode: 1,
		},
		{
			name:        "not active",
			sessionName: "gamma",
			setup: func(t *testing.T) {
				createDesktopSessionFixture(t, desktopSessionFixture{
					Name:    "gamma",
					IP:      "127.0.0.1",
					Pid:     999999,
					State:   "active",
					Host:    "localhost",
					Display: "3",
				})
			},
			expectedResponse: map[string]any{
				"success":      false,
				"session_name": "gamma",
				"error":        "Desktop session 'gamma' is not active.",
				"reason":       "not_active",
			},
			expectedExitCode: 1,
		},
		{
			name:        "not found",
			sessionName: "delta",
			setup:       func(t *testing.T) {},
			expectedResponse: map[string]any{
				"success":      false,
				"session_name": "delta",
				"error":        "Unknown session: delta",
				"reason":       "not_found",
			},
			expectedExitCode: 1,
		},
		{
			name:        "terminate failed",
			sessionName: "epsilon",
			setup: func(t *testing.T) {
				proc := startTestProcess(t)
				createDesktopSessionFixture(t, desktopSessionFixture{
					Name:    "epsilon",
					IP:      "127.0.0.1",
					Pid:     proc.Process.Pid,
					State:   "active",
					Host:    "localhost",
					Display: "4",
				})
				overrideTestVNCServer(t, "#!/bin/sh\nexit 1\n")
			},
			expectedResponse: map[string]any{
				"success":      false,
				"session_name": "epsilon",
				"error":        "Terminating session 'epsilon' failed.",
				"reason":       "terminate_failed",
			},
			expectedExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			output, err := runBinary([]string{"kill", tt.sessionName, "--format", "json"}, nil)
			assertExitCode(t, tt.expectedExitCode, output, err)

			var got map[string]any
			if err := json.NewDecoder(bytes.NewReader(output)).Decode(&got); err != nil {
				t.Fatalf("failed to decode JSON output: %v\noutput: %s", err, output)
			}

			if len(got) != len(tt.expectedResponse) {
				t.Fatalf("expected %d JSON fields, got %d: %#v", len(tt.expectedResponse), len(got), got)
			}
			for key, want := range tt.expectedResponse {
				if got[key] != want {
					t.Fatalf("expected %s=%v, got %v", key, want, got[key])
				}
			}
		})
	}
}

type desktopSessionFixture struct {
	Name    string
	IP      string
	Pid     int
	State   string
	Host    string
	Display string
}

func createDesktopSessionFixture(t *testing.T, session desktopSessionFixture) {
	t.Helper()

	sessionDir := filepath.Join(tmpDir, "local", "state", "flight", "desktop", "sessions", session.Name)
	t.Cleanup(func() {
		_ = os.RemoveAll(sessionDir)
	})
	if err := os.MkdirAll(sessionDir, 0o700); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	pidfile := filepath.Join(sessionDir, "vncserver.pid")
	if err := os.WriteFile(pidfile, fmt.Appendf(nil, "%d\n", session.Pid), 0o600); err != nil {
		t.Fatalf("failed to write pidfile: %v", err)
	}

	metadata := fmt.Sprintf(`session_type: xterm
password: hunter2
ip: %s
state: %s
created_at: %s
metadata:
  host: %s
  display: "%s"
  pidfile: %s
`, session.IP, session.State, time.Date(2026, time.April, 22, 10, 0, 0, 0, time.UTC).Format(time.RFC3339), session.Host, session.Display, pidfile)

	if err := os.WriteFile(filepath.Join(sessionDir, "metadata.yml"), []byte(metadata), 0o600); err != nil {
		t.Fatalf("failed to write metadata: %v", err)
	}
}

// We need a process running that can (1) be used to determine that a session
// is active and not exited; and (2) be killed by the `kill` command. We create
// such a process here.
func startTestProcess(t *testing.T) *exec.Cmd {
	t.Helper()

	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper process: %v", err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})
	return cmd
}

// overrideTestVNCServer overrides vncserver for a single test. This allows
// control over whether the execution of it fails or not.  For instance, some
// tests want to simulate a failure to kill a desktop session.
func overrideTestVNCServer(t *testing.T, script string) {
	t.Helper()

	path := filepath.Join(flightRoot, "usr", "libexec", "desktop", "vncserver")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read vncserver fixture: %v", err)
	}
	t.Cleanup(func() {
		if err := os.WriteFile(path, original, 0o755); err != nil {
			t.Fatalf("failed to restore vncserver fixture: %v", err)
		}
	})
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to override vncserver fixture: %v", err)
	}
}
