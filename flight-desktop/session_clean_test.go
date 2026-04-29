package main_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type cleanCommandResponse struct {
	Success bool                 `json:"success"`
	Results []cleanCommandResult `json:"results"`
}

type cleanCommandResult struct {
	Success     bool   `json:"success"`
	SessionName string `json:"session_name"`
	Error       string `json:"error,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

func TestCleanSessionJSON(t *testing.T) {
	tests := []struct {
		name             string
		sessionName      string
		setup            func(t *testing.T)
		expectedResponse cleanCommandResponse
		expectedExitCode int
		expectRemoved    bool
	}{
		{
			name:        "success",
			sessionName: "alpha",
			setup: func(t *testing.T) {
				createDesktopSessionFixture(t, desktopSessionFixture{
					Name:    "alpha",
					IP:      "127.0.0.1",
					Pid:     999999,
					State:   "exited",
					Host:    "localhost",
					Display: "1",
				})
			},
			expectedResponse: cleanCommandResponse{
				Success: true,
				Results: []cleanCommandResult{
					{
						Success:     true,
						SessionName: "alpha",
					},
				},
			},
			expectedExitCode: 0,
			expectRemoved:    true,
		},
		{
			name:        "not local",
			sessionName: "beta",
			setup: func(t *testing.T) {
				createDesktopSessionFixture(t, desktopSessionFixture{
					Name:    "beta",
					IP:      "192.0.2.10",
					Pid:     999999,
					State:   "exited",
					Host:    "remote-host",
					Display: "2",
				})
			},
			expectedResponse: cleanCommandResponse{
				Success: false,
				Results: []cleanCommandResult{
					{
						Success:     false,
						SessionName: "beta",
						Error:       "Desktop session 'beta' is not local.",
						Reason:      "not_local",
					},
				},
			},
			expectedExitCode: 1,
		},
		{
			name:        "active",
			sessionName: "gamma",
			setup: func(t *testing.T) {
				proc := startTestProcess(t)
				createDesktopSessionFixture(t, desktopSessionFixture{
					Name:    "gamma",
					IP:      "127.0.0.1",
					Pid:     proc.Process.Pid,
					State:   "active",
					Host:    "localhost",
					Display: "3",
				})
			},
			expectedResponse: cleanCommandResponse{
				Success: false,
				Results: []cleanCommandResult{
					{
						Success:     false,
						SessionName: "gamma",
						Error:       "Desktop session 'gamma' is active.",
						Reason:      "active",
					},
				},
			},
			expectedExitCode: 1,
		},
		{
			name:        "not found treated as success",
			sessionName: "delta",
			setup:       func(t *testing.T) {},
			expectedResponse: cleanCommandResponse{
				Success: true,
				Results: []cleanCommandResult{
					{
						Success:     true,
						SessionName: "delta",
					},
				},
			},
			expectedExitCode: 0,
		},
		{
			name:        "clean failed",
			sessionName: "epsilon",
			setup: func(t *testing.T) {
				createDesktopSessionFixture(t, desktopSessionFixture{
					Name:    "epsilon",
					IP:      "127.0.0.1",
					Pid:     999999,
					State:   "exited",
					Host:    "localhost",
					Display: "4",
				})
				sessionsDir := filepath.Join(tmpDir, "local", "state", "flight", "desktop", "sessions")
				if err := os.Chmod(sessionsDir, 0o500); err != nil {
					t.Fatalf("failed to make sessions dir read-only: %v", err)
				}
				t.Cleanup(func() {
					_ = os.Chmod(sessionsDir, 0o755)
				})
			},
			expectedResponse: cleanCommandResponse{
				Success: false,
				Results: []cleanCommandResult{
					{
						Success:     false,
						SessionName: "epsilon",
						Error:       "Cleaning session 'epsilon' failed.",
						Reason:      "clean_failed",
					},
				},
			},
			expectedExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetDesktopSessionState(t)
			tt.setup(t)

			output, err := runBinary([]string{"clean", "--format", "json", "--", tt.sessionName}, nil)
			assertExitCode(t, tt.expectedExitCode, output, err)

			var got cleanCommandResponse
			if err := json.NewDecoder(bytes.NewReader(output)).Decode(&got); err != nil {
				t.Fatalf("failed to decode JSON output: %v\noutput: %s", err, output)
			}

			if got.Success != tt.expectedResponse.Success {
				t.Fatalf("expected success=%v, got %v", tt.expectedResponse.Success, got.Success)
			}
			if len(got.Results) != len(tt.expectedResponse.Results) {
				t.Fatalf("expected %d results, got %d: %#v", len(tt.expectedResponse.Results), len(got.Results), got.Results)
			}
			for i, want := range tt.expectedResponse.Results {
				gotResult := got.Results[i]
				if gotResult != want {
					t.Fatalf("expected result[%d]=%#v, got %#v", i, want, gotResult)
				}
			}

			sessionDir := filepath.Join(tmpDir, "local", "state", "flight", "desktop", "sessions", tt.sessionName)
			_, statErr := os.Stat(sessionDir)
			if tt.expectRemoved && !os.IsNotExist(statErr) {
				t.Fatalf("expected session dir %q to be removed, stat err=%v", sessionDir, statErr)
			}
		})
	}
}
