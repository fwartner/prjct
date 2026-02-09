package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func setReadmeOutput(t *testing.T, val string) {
	t.Helper()
	old := readmeOutput
	readmeOutput = val
	t.Cleanup(func() { readmeOutput = old })
}

func TestRunReadme(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runReadme(cmd, []string{"test"})
	if err != nil {
		t.Fatalf("runReadme() error: %v", err)
	}
}

func TestRunReadmeToFile(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	outputPath := filepath.Join(t.TempDir(), "README.md")
	setReadmeOutput(t, outputPath)

	cmd := &cobra.Command{}
	err := runReadme(cmd, []string{"test"})
	if err != nil {
		t.Fatalf("runReadme() error: %v", err)
	}

	data, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatalf("read output: %v", readErr)
	}

	content := string(data)
	if !strings.Contains(content, "# Test Template") {
		t.Error("README should contain template name as heading")
	}
	if !strings.Contains(content, "src/") {
		t.Error("README should list directories")
	}
}

func TestRunReadmeNotFound(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runReadme(cmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing template")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitTemplateNotFound {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitTemplateNotFound)
	}
}
