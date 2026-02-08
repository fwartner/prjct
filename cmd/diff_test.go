package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunDiff(t *testing.T) {
	// Create config with a template
	dir := t.TempDir()
	content := `templates:
  - id: video
    name: "Video Production"
    base_path: "/tmp"
    directories:
      - name: "Pre-Production"
      - name: "Production"
      - name: "Post"
`
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, cfgPath)

	// Create a project directory with some matching and extra dirs
	projectDir := filepath.Join(dir, "myproject")
	for _, d := range []string{"Pre-Production", "Production", "Extra"} {
		if err := os.MkdirAll(filepath.Join(projectDir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	cmd := &cobra.Command{}
	err := runDiff(cmd, []string{"video", projectDir})
	if err != nil {
		t.Fatalf("runDiff() error: %v", err)
	}
}

func TestRunDiffAllMatching(t *testing.T) {
	dir := t.TempDir()
	content := `templates:
  - id: simple
    name: "Simple"
    base_path: "/tmp"
    directories:
      - name: "src"
      - name: "docs"
`
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, cfgPath)

	projectDir := filepath.Join(dir, "proj")
	for _, d := range []string{"src", "docs"} {
		if err := os.MkdirAll(filepath.Join(projectDir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	cmd := &cobra.Command{}
	err := runDiff(cmd, []string{"simple", projectDir})
	if err != nil {
		t.Fatalf("runDiff() all matching error: %v", err)
	}
}

func TestRunDiffTemplateNotFound(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runDiff(cmd, []string{"nonexistent", "/tmp"})
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitTemplateNotFound {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitTemplateNotFound)
	}
}

func TestRunDiffProjectNotFound(t *testing.T) {
	dir := t.TempDir()
	content := `templates:
  - id: test
    name: "Test"
    base_path: "/tmp"
    directories:
      - name: "src"
`
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runDiff(cmd, []string{"test", "/nonexistent/path"})
	if err == nil {
		t.Fatal("expected error for missing project path")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitGeneral {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitGeneral)
	}
}
