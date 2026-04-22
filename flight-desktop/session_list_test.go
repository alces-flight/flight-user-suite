package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestWriteSessionsJSON(t *testing.T) {
	sessions := []*Session{
		{
			Name:        "alpha",
			SessionType: "gnome",
			State:       Active,
			CreatedAt:   time.Date(2026, time.April, 21, 11, 30, 0, 0, time.UTC),
			Metadata: sessionMetadata{
				Host: "desktop-a",
			},
		},
	}

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = writer
	t.Cleanup(func() {
		os.Stdout = originalStdout
	})

	if err := writeSessionsJSON(sessions); err != nil {
		t.Fatalf("writeSessionsJSON returned error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	var got []listedSession
	if err := json.NewDecoder(reader).Decode(&got); err != nil {
		t.Fatalf("failed to decode JSON output: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 session, got %d", len(got))
	}

	session := got[0]
	if session.Name != "alpha" {
		t.Errorf("expected name alpha, got %q", session.Name)
	}
	if session.DesktopType != "gnome" {
		t.Errorf("expected desktop type gnome, got %q", session.DesktopType)
	}
	if session.State != Remote {
		t.Errorf("expected state %q, got %q", Remote, session.State)
	}
	if session.Host != "desktop-a" {
		t.Errorf("expected host desktop-a, got %q", session.Host)
	}
	if session.CreatedAt != "2026-04-21T11:30:00Z" {
		t.Errorf("expected created at 2026-04-21T11:30:00Z, got %q", session.CreatedAt)
	}
}
