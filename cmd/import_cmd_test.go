package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunImport(t *testing.T) {
	dir := t.TempDir()

	// Existing config with one template
	cfgContent := `templates:
  - id: existing
    name: "Existing"
    base_path: "/tmp"
    directories:
      - name: "src"
`
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, cfgPath)

	// Import file with a new template
	importContent := `templates:
  - id: imported
    name: "Imported Template"
    base_path: "/tmp"
    directories:
      - name: "docs"
`
	importPath := filepath.Join(dir, "import.yaml")
	if err := os.WriteFile(importPath, []byte(importContent), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{}
	err := runImport(cmd, []string{importPath})
	if err != nil {
		t.Fatalf("runImport() error: %v", err)
	}

	// Verify the config now has both templates
	data, readErr := os.ReadFile(cfgPath)
	if readErr != nil {
		t.Fatalf("reading config: %v", readErr)
	}
	content := string(data)
	if !(contains(content, "existing") && contains(content, "imported")) {
		t.Error("config should contain both existing and imported templates")
	}
}

func TestRunImportSkipsConflict(t *testing.T) {
	dir := t.TempDir()

	cfgContent := `templates:
  - id: video
    name: "Video"
    base_path: "/tmp"
    directories:
      - name: "src"
`
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, cfgPath)

	// Import file with same ID
	importContent := `templates:
  - id: video
    name: "Different Video"
    base_path: "/tmp"
    directories:
      - name: "new"
`
	importPath := filepath.Join(dir, "import.yaml")
	if err := os.WriteFile(importPath, []byte(importContent), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{}
	err := runImport(cmd, []string{importPath})
	if err != nil {
		t.Fatalf("runImport() skip error: %v", err)
	}
}

func TestRunImportEmptyFile(t *testing.T) {
	dir := t.TempDir()

	cfgContent := `templates:
  - id: test
    name: "Test"
    base_path: "/tmp"
    directories:
      - name: "src"
`
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, cfgPath)

	importContent := `templates: []
`
	importPath := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(importPath, []byte(importContent), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{}
	err := runImport(cmd, []string{importPath})
	if err == nil {
		t.Fatal("expected error for empty import file")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
}

func TestRunImportBadFile(t *testing.T) {
	dir := t.TempDir()

	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("templates: []"), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runImport(cmd, []string{"/nonexistent/file.yaml"})
	if err == nil {
		t.Fatal("expected error for missing import file")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
