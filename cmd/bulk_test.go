package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunBulk(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Write manifest
	manifest := `projects:
  - template: test
    name: "BulkProject1"
  - template: test
    name: "BulkProject2"
`
	manifestPath := filepath.Join(t.TempDir(), "manifest.yaml")
	_ = os.WriteFile(manifestPath, []byte(manifest), 0644)

	cmd := &cobra.Command{}
	err := runBulk(cmd, []string{manifestPath})
	if err != nil {
		t.Fatalf("runBulk() error: %v", err)
	}

	// Verify both projects created
	for _, name := range []string{"BulkProject1", "BulkProject2"} {
		p := filepath.Join(base, name)
		if _, statErr := os.Stat(p); os.IsNotExist(statErr) {
			t.Errorf("project %q not created", name)
		}
	}
}

func TestRunBulkBadTemplate(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	manifest := `projects:
  - template: nonexistent
    name: "WontWork"
`
	manifestPath := filepath.Join(t.TempDir(), "manifest.yaml")
	_ = os.WriteFile(manifestPath, []byte(manifest), 0644)

	cmd := &cobra.Command{}
	// Should not return error — individual failures are reported but don't stop
	err := runBulk(cmd, []string{manifestPath})
	if err != nil {
		t.Fatalf("runBulk() should not error: %v", err)
	}
}

func TestRunBulkEmptyManifest(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	manifest := `projects: []`
	manifestPath := filepath.Join(t.TempDir(), "manifest.yaml")
	_ = os.WriteFile(manifestPath, []byte(manifest), 0644)

	cmd := &cobra.Command{}
	err := runBulk(cmd, []string{manifestPath})
	if err == nil {
		t.Fatal("expected error for empty manifest")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
}

func TestRunBulkBadFile(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runBulk(cmd, []string{"/nonexistent/manifest.yaml"})
	if err == nil {
		t.Fatal("expected error for missing manifest")
	}
}
