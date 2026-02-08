package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func setExportOutput(t *testing.T, val string) {
	t.Helper()
	old := exportOutput
	exportOutput = val
	t.Cleanup(func() { exportOutput = old })
}

func TestRunExport(t *testing.T) {
	dir := t.TempDir()
	content := `templates:
  - id: video
    name: "Video Production"
    base_path: "/tmp"
    directories:
      - name: "src"
`
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, cfgPath)

	outPath := filepath.Join(dir, "exported.yaml")
	setExportOutput(t, outPath)

	cmd := &cobra.Command{}
	err := runExport(cmd, []string{"video"})
	if err != nil {
		t.Fatalf("runExport() error: %v", err)
	}

	data, readErr := os.ReadFile(outPath)
	if readErr != nil {
		t.Fatalf("reading exported file: %v", readErr)
	}
	if !strings.Contains(string(data), "video") {
		t.Error("exported file should contain template ID")
	}
}

func TestRunExportDefaultOutput(t *testing.T) {
	dir := t.TempDir()
	content := `templates:
  - id: mytemplate
    name: "My Template"
    base_path: "/tmp"
    directories:
      - name: "src"
`
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, cfgPath)
	setExportOutput(t, "")

	// Change to temp dir so default output goes there
	oldWd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(oldWd) })

	cmd := &cobra.Command{}
	err := runExport(cmd, []string{"mytemplate"})
	if err != nil {
		t.Fatalf("runExport() default output error: %v", err)
	}

	defaultPath := filepath.Join(dir, "mytemplate.yaml")
	if _, statErr := os.Stat(defaultPath); os.IsNotExist(statErr) {
		t.Error("default output file was not created")
	}
}

func TestRunExportNotFound(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)
	setExportOutput(t, "")

	cmd := &cobra.Command{}
	err := runExport(cmd, []string{"nonexistent"})
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
