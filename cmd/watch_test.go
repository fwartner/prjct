package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fwartner/prjct/internal/config"
	"github.com/fwartner/prjct/internal/index"
)

func TestScanAndIndex(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Create a directory in the base path that looks like a project
	_ = os.MkdirAll(filepath.Join(base, "NewProject"), 0755)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	cfg := &config.Config{
		Templates: []config.Template{
			{ID: "test", Name: "Test", BasePath: base},
		},
	}

	scanAndIndex(cfg, idxPath)

	idx, err := index.Load(idxPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(idx.Projects) == 0 {
		t.Fatal("expected at least 1 project indexed")
	}

	// Scan again — should not duplicate
	scanAndIndex(cfg, idxPath)

	idx2, _ := index.Load(idxPath)
	if len(idx2.Projects) != len(idx.Projects) {
		t.Errorf("duplicate indexing: first=%d second=%d", len(idx.Projects), len(idx2.Projects))
	}
}

func TestScanAndIndexSkipsExisting(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	projectDir := filepath.Join(base, "Existing")
	_ = os.MkdirAll(projectDir, 0755)

	// Pre-populate index
	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	preIdx := &index.Index{
		Projects: []index.Entry{
			{Name: "Existing", TemplateID: "test", Path: projectDir, CreatedAt: time.Now()},
		},
	}
	data, _ := json.MarshalIndent(preIdx, "", "  ")
	_ = os.WriteFile(idxPath, data, 0644)

	cfg := &config.Config{
		Templates: []config.Template{
			{ID: "test", Name: "Test", BasePath: base},
		},
	}

	scanAndIndex(cfg, idxPath)

	idx, _ := index.Load(idxPath)
	count := 0
	for _, e := range idx.Projects {
		if e.Path == projectDir {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 entry for Existing, got %d", count)
	}
}
