package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunTree(t *testing.T) {
	dir := t.TempDir()
	content := `templates:
  - id: video
    name: "Video Production"
    base_path: "/tmp"
    directories:
      - name: "Pre-Production"
        children:
          - name: "Scripts"
      - name: "Production"
`
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runTree(cmd, []string{"video"})
	if err != nil {
		t.Fatalf("runTree() error: %v", err)
	}
}

func TestRunTreeNotFound(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runTree(cmd, []string{"nonexistent"})
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

func TestRunTreeWithOptionalAndFiles(t *testing.T) {
	dir := t.TempDir()
	content := `templates:
  - id: test
    name: "Test"
    base_path: "/tmp"
    directories:
      - name: "src"
        files:
          - name: "main.go"
            content: "package main"
      - name: "optional"
        optional: true
`
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runTree(cmd, []string{"test"})
	if err != nil {
		t.Fatalf("runTree() error: %v", err)
	}
}
