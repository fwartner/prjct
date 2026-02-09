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

func TestRunSync(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Create a project with only "src" (missing "docs")
	projectDir := filepath.Join(base, "SyncTest")
	if err := os.MkdirAll(filepath.Join(projectDir, "src"), 0755); err != nil {
		t.Fatal(err)
	}

	// Add to index
	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	idx := &index.Index{
		Projects: []index.Entry{
			{Name: "SyncTest", TemplateID: "test", TemplateName: "Test Template", Path: projectDir, CreatedAt: time.Now()},
		},
	}
	data, _ := json.MarshalIndent(idx, "", "  ")
	_ = os.WriteFile(idxPath, data, 0644)

	cmd := &cobra.Command{}
	err := runSync(cmd, []string{"SyncTest"})
	if err != nil {
		t.Fatalf("runSync() error: %v", err)
	}

	// Check that "docs" was created
	docsPath := filepath.Join(projectDir, "docs")
	if _, statErr := os.Stat(docsPath); os.IsNotExist(statErr) {
		t.Error("expected docs directory to be created by sync")
	}
}

func TestRunSyncAlreadyInSync(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Create a complete project
	projectDir := filepath.Join(base, "Complete")
	_ = os.MkdirAll(filepath.Join(projectDir, "src"), 0755)
	_ = os.MkdirAll(filepath.Join(projectDir, "docs"), 0755)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	idx := &index.Index{
		Projects: []index.Entry{
			{Name: "Complete", TemplateID: "test", Path: projectDir, CreatedAt: time.Now()},
		},
	}
	data, _ := json.MarshalIndent(idx, "", "  ")
	_ = os.WriteFile(idxPath, data, 0644)

	cmd := &cobra.Command{}
	err := runSync(cmd, []string{"Complete"})
	if err != nil {
		t.Fatalf("runSync() error: %v", err)
	}
}

func TestRunSyncNoMatch(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Empty index
	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	_ = os.WriteFile(idxPath, []byte(`{"projects":[]}`), 0644)

	cmd := &cobra.Command{}
	err := runSync(cmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for no match")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
}
