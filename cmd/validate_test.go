package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunValidateValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	content := `templates:
  - id: mytemplate
    name: "My Template"
    base_path: "/tmp"
    directories:
      - name: "src"
`
	_ = os.WriteFile(path, []byte(content), 0644)

	cmd := &cobra.Command{}
	err := runValidate(cmd, []string{path})
	if err != nil {
		t.Fatalf("runValidate() error: %v", err)
	}
}

func TestRunValidateInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	content := `templates:
  - id: ""
    name: ""
    base_path: ""
`
	_ = os.WriteFile(path, []byte(content), 0644)

	cmd := &cobra.Command{}
	err := runValidate(cmd, []string{path})
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitConfigInvalid {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitConfigInvalid)
	}
}

func TestRunValidateBadYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "corrupt.yaml")
	_ = os.WriteFile(path, []byte(":::bad:::"), 0644)

	cmd := &cobra.Command{}
	err := runValidate(cmd, []string{path})
	if err == nil {
		t.Fatal("expected error for bad YAML")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
}

func TestRunValidateNotFound(t *testing.T) {
	cmd := &cobra.Command{}
	err := runValidate(cmd, []string{"/nonexistent/file.yaml"})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
