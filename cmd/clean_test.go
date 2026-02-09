package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

func TestRunClean(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Create project with empty dirs
	projectDir := filepath.Join(base, "Messy")
	os.MkdirAll(filepath.Join(projectDir, "src"), 0755)
	os.MkdirAll(filepath.Join(projectDir, "empty1"), 0755)
	os.MkdirAll(filepath.Join(projectDir, "empty2", "nested"), 0755)
	// src has a file so it won't be removed
	_ = os.WriteFile(filepath.Join(projectDir, "src", "main.go"), []byte("x"), 0644)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	idx := &index.Index{
		Projects: []index.Entry{
			{Name: "Messy", TemplateID: "test", Path: projectDir, CreatedAt: time.Now()},
		},
	}
	data, _ := json.MarshalIndent(idx, "", "  ")
	_ = os.WriteFile(idxPath, data, 0644)

	cmd := &cobra.Command{}
	err := runClean(cmd, []string{"Messy"})
	if err != nil {
		t.Fatalf("runClean() error: %v", err)
	}

	// empty1 should be removed
	if _, statErr := os.Stat(filepath.Join(projectDir, "empty1")); !os.IsNotExist(statErr) {
		t.Error("expected empty1 to be removed")
	}
	// nested should be removed first, then empty2
	if _, statErr := os.Stat(filepath.Join(projectDir, "empty2")); !os.IsNotExist(statErr) {
		t.Error("expected empty2 to be removed")
	}
	// src should remain (has a file)
	if _, statErr := os.Stat(filepath.Join(projectDir, "src")); os.IsNotExist(statErr) {
		t.Error("src should remain (has files)")
	}
}

func TestRunCleanNoEmpty(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	projectDir := filepath.Join(base, "Full")
	os.MkdirAll(filepath.Join(projectDir, "src"), 0755)
	_ = os.WriteFile(filepath.Join(projectDir, "src", "a.txt"), []byte("x"), 0644)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	idx := &index.Index{
		Projects: []index.Entry{
			{Name: "Full", TemplateID: "test", Path: projectDir, CreatedAt: time.Now()},
		},
	}
	data, _ := json.MarshalIndent(idx, "", "  ")
	_ = os.WriteFile(idxPath, data, 0644)

	cmd := &cobra.Command{}
	err := runClean(cmd, []string{"Full"})
	if err != nil {
		t.Fatalf("runClean() error: %v", err)
	}
}
