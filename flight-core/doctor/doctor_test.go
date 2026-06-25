package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunPythonModuleCheck(t *testing.T) {
	tmpDir := t.TempDir()
	pythonPath := filepath.Join(tmpDir, "python3")
	script := "#!/bin/sh\ncase \"$2\" in\n  \"import pam\") exit 0 ;;\n  *) echo \"unexpected invocation: $*\" >&2; exit 2 ;;\nesac\n"
	if err := os.WriteFile(pythonPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	origPath := os.Getenv("PATH")
	t.Cleanup(func() {
		_ = os.Setenv("PATH", origPath)
	})
	if err := os.Setenv("PATH", tmpDir); err != nil {
		t.Fatal(err)
	}

	results, ok := Run([]Dependency{{
		Type:   TypePythonModule,
		Paths:  []string{"python3"},
		Module: "pam",
	}})
	if !ok {
		t.Fatalf("expected python module check to pass, got %#v", results)
	}
	if len(results) != 1 || !results[0].Found {
		t.Fatalf("expected single passing result, got %#v", results)
	}
}

func TestRunPythonModuleCheckFailure(t *testing.T) {
	tmpDir := t.TempDir()
	pythonPath := filepath.Join(tmpDir, "python3")
	script := "#!/bin/sh\necho \"ModuleNotFoundError: No module named 'pam'\" >&2\nexit 1\n"
	if err := os.WriteFile(pythonPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	origPath := os.Getenv("PATH")
	t.Cleanup(func() {
		_ = os.Setenv("PATH", origPath)
	})
	if err := os.Setenv("PATH", tmpDir); err != nil {
		t.Fatal(err)
	}

	results, ok := Run([]Dependency{{
		Type:   TypePythonModule,
		Paths:  []string{"python3"},
		Module: "pam",
	}})
	if ok {
		t.Fatalf("expected python module check to fail, got %#v", results)
	}
	if len(results) != 1 || results[0].Err == nil {
		t.Fatalf("expected single failing result, got %#v", results)
	}
}
