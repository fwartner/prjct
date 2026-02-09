package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

func setCloneWithFiles(t *testing.T, val bool) {
	t.Helper()
	old := cloneWithFiles
	cloneWithFiles = val
	t.Cleanup(func() { cloneWithFiles = old })
}

func TestRunClone(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Create source project
	sourceDir := filepath.Join(base, "Original")
	_ = os.MkdirAll(filepath.Join(sourceDir, "src"), 0755)
	_ = os.MkdirAll(filepath.Join(sourceDir, "docs"), 0755)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	idx := &index.Index{
		Projects: []index.Entry{
			{Name: "Original", TemplateID: "test", TemplateName: "Test Template", Path: sourceDir, CreatedAt: time.Now()},
		},
	}
	data, _ := json.MarshalIndent(idx, "", "  ")
	_ = os.WriteFile(idxPath, data, 0644)

	cmd := &cobra.Command{}
	err := runClone(cmd, []string{"Original", "Cloned"})
	if err != nil {
		t.Fatalf("runClone() error: %v", err)
	}

	// Verify clone exists
	clonedDir := filepath.Join(base, "Cloned")
	if _, statErr := os.Stat(clonedDir); os.IsNotExist(statErr) {
		t.Error("cloned directory not created")
	}
	if _, statErr := os.Stat(filepath.Join(clonedDir, "src")); os.IsNotExist(statErr) {
		t.Error("cloned src not created")
	}
}

func TestRunCloneWithFiles(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)
	setCloneWithFiles(t, true)

	// Create source with a file
	sourceDir := filepath.Join(base, "Source")
	_ = os.MkdirAll(filepath.Join(sourceDir, "src"), 0755)
	_ = os.WriteFile(filepath.Join(sourceDir, "src", "main.go"), []byte("package main"), 0644)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	idx := &index.Index{
		Projects: []index.Entry{
			{Name: "Source", TemplateID: "test", Path: sourceDir, CreatedAt: time.Now()},
		},
	}
	data, _ := json.MarshalIndent(idx, "", "  ")
	_ = os.WriteFile(idxPath, data, 0644)

	cmd := &cobra.Command{}
	err := runClone(cmd, []string{"Source", "Copy"})
	if err != nil {
		t.Fatalf("runClone() error: %v", err)
	}

	// Verify file was copied
	copiedFile := filepath.Join(base, "Copy", "src", "main.go")
	data, readErr := os.ReadFile(copiedFile)
	if readErr != nil {
		t.Fatalf("file not copied: %v", readErr)
	}
	if string(data) != "package main" {
		t.Errorf("file content = %q, want %q", string(data), "package main")
	}
}

func TestRunCloneNoMatch(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	_ = os.WriteFile(idxPath, []byte(`{"projects":[]}`), 0644)

	cmd := &cobra.Command{}
	err := runClone(cmd, []string{"nope", "Clone"})
	if err == nil {
		t.Fatal("expected error")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
}

func TestRunCloneDestExists(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	sourceDir := filepath.Join(base, "Src")
	_ = os.MkdirAll(sourceDir, 0755)
	_ = os.MkdirAll(filepath.Join(base, "Exists"), 0755) // dest exists

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	idx := &index.Index{
		Projects: []index.Entry{
			{Name: "Src", TemplateID: "test", Path: sourceDir, CreatedAt: time.Now()},
		},
	}
	data, _ := json.MarshalIndent(idx, "", "  ")
	_ = os.WriteFile(idxPath, data, 0644)

	cmd := &cobra.Command{}
	err := runClone(cmd, []string{"Src", "Exists"})
	if err == nil {
		t.Fatal("expected error for existing dest")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitProjectExists {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitProjectExists)
	}
}
