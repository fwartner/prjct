package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

// setReindexTemplate sets the package-level reindexTemplate for the duration of the test.
func setReindexTemplate(t *testing.T, val string) {
	t.Helper()
	old := reindexTemplate
	reindexTemplate = val
	t.Cleanup(func() { reindexTemplate = old })
}

// setupReindexEnv creates a config and project directories, returns the config dir.
func setupReindexEnv(t *testing.T) (cfgDir string, base string) {
	t.Helper()
	base = t.TempDir()
	cfgDir = t.TempDir()

	// Create some "project" directories under base
	for _, name := range []string{"ProjectA", "ProjectB", "ProjectC"} {
		if err := os.Mkdir(filepath.Join(base, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	// Create a file (should be ignored, not a directory)
	if err := os.WriteFile(filepath.Join(base, "notes.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	// Write config pointing to base
	content := fmt.Sprintf(`templates:
  - id: test
    name: "Test Template"
    base_path: %q
    directories:
      - name: "src"
`, base)
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, cfgPath)
	return cfgDir, base
}

func TestRunReindexDiscoversProjects(t *testing.T) {
	cfgDir, _ := setupReindexEnv(t)
	setReindexTemplate(t, "")
	setVerbose(t, false)

	cmd := &cobra.Command{}
	err := runReindex(cmd, nil)
	if err != nil {
		t.Fatalf("runReindex() error: %v", err)
	}

	// Verify index was created with 3 projects
	idxPath := filepath.Join(cfgDir, "projects.json")
	idx, loadErr := index.Load(idxPath)
	if loadErr != nil {
		t.Fatalf("Load() error: %v", loadErr)
	}
	if len(idx.Projects) != 3 {
		t.Errorf("indexed %d projects, want 3", len(idx.Projects))
	}
}

func TestRunReindexSkipsDuplicates(t *testing.T) {
	cfgDir, base := setupReindexEnv(t)
	setReindexTemplate(t, "")
	setVerbose(t, false)

	// Pre-populate index with one entry
	idxPath := filepath.Join(cfgDir, "projects.json")
	preIdx := &index.Index{
		Projects: []index.Entry{
			{Name: "ProjectA", TemplateID: "test", Path: filepath.Join(base, "ProjectA")},
		},
	}
	data, _ := json.MarshalIndent(preIdx, "", "  ")
	if err := os.WriteFile(idxPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{}
	err := runReindex(cmd, nil)
	if err != nil {
		t.Fatalf("runReindex() error: %v", err)
	}

	idx, _ := index.Load(idxPath)
	if len(idx.Projects) != 3 {
		t.Errorf("indexed %d projects, want 3 (1 existing + 2 new)", len(idx.Projects))
	}
}

func TestRunReindexWithTemplateFilter(t *testing.T) {
	base := t.TempDir()
	base2 := t.TempDir()
	cfgDir := t.TempDir()

	// Create projects in both base paths
	os.Mkdir(filepath.Join(base, "ProjA"), 0755)
	os.Mkdir(filepath.Join(base2, "ProjB"), 0755)

	content := fmt.Sprintf(`templates:
  - id: video
    name: "Video"
    base_path: %q
    directories:
      - name: "src"
  - id: photo
    name: "Photo"
    base_path: %q
    directories:
      - name: "raw"
`, base, base2)
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	os.WriteFile(cfgPath, []byte(content), 0644)
	setConfigPath(t, cfgPath)
	setReindexTemplate(t, "video")
	setVerbose(t, false)

	cmd := &cobra.Command{}
	err := runReindex(cmd, nil)
	if err != nil {
		t.Fatalf("runReindex() error: %v", err)
	}

	idxPath := filepath.Join(cfgDir, "projects.json")
	idx, _ := index.Load(idxPath)
	if len(idx.Projects) != 1 {
		t.Errorf("indexed %d projects, want 1 (only video)", len(idx.Projects))
	}
	if len(idx.Projects) > 0 && idx.Projects[0].TemplateID != "video" {
		t.Errorf("TemplateID = %q, want video", idx.Projects[0].TemplateID)
	}
}

func TestRunReindexEmptyBasePath(t *testing.T) {
	base := t.TempDir() // empty directory
	cfgDir := t.TempDir()

	content := fmt.Sprintf(`templates:
  - id: test
    name: "Test"
    base_path: %q
    directories:
      - name: "src"
`, base)
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	os.WriteFile(cfgPath, []byte(content), 0644)
	setConfigPath(t, cfgPath)
	setReindexTemplate(t, "")
	setVerbose(t, false)

	cmd := &cobra.Command{}
	err := runReindex(cmd, nil)
	if err != nil {
		t.Fatalf("runReindex() error: %v", err)
	}

	idxPath := filepath.Join(cfgDir, "projects.json")
	idx, _ := index.Load(idxPath)
	if len(idx.Projects) != 0 {
		t.Errorf("indexed %d projects from empty dir, want 0", len(idx.Projects))
	}
}

func TestRunReindexNonexistentBasePath(t *testing.T) {
	cfgDir := t.TempDir()

	content := `templates:
  - id: test
    name: "Test"
    base_path: "/nonexistent/path/nowhere"
    directories:
      - name: "src"
`
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	os.WriteFile(cfgPath, []byte(content), 0644)
	setConfigPath(t, cfgPath)
	setReindexTemplate(t, "")
	setVerbose(t, false)

	cmd := &cobra.Command{}
	// Should succeed (skip unreadable base paths)
	err := runReindex(cmd, nil)
	if err != nil {
		t.Fatalf("runReindex() error: %v", err)
	}
}

func TestRunReindexVerbose(t *testing.T) {
	cfgDir, _ := setupReindexEnv(t)
	setReindexTemplate(t, "")
	setVerbose(t, true)

	cmd := &cobra.Command{}
	err := runReindex(cmd, nil)
	if err != nil {
		t.Fatalf("runReindex() verbose error: %v", err)
	}

	idxPath := filepath.Join(cfgDir, "projects.json")
	idx, _ := index.Load(idxPath)
	if len(idx.Projects) != 3 {
		t.Errorf("indexed %d projects, want 3", len(idx.Projects))
	}
}

func TestRunReindexTemplateNotFound(t *testing.T) {
	cfgDir := t.TempDir()
	base := t.TempDir()

	content := fmt.Sprintf(`templates:
  - id: test
    name: "Test"
    base_path: %q
    directories:
      - name: "src"
`, base)
	cfgPath := filepath.Join(cfgDir, "config.yaml")
	os.WriteFile(cfgPath, []byte(content), 0644)
	setConfigPath(t, cfgPath)
	setReindexTemplate(t, "nonexistent")
	setVerbose(t, false)

	cmd := &cobra.Command{}
	err := runReindex(cmd, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent template filter")
	}
}

func TestRunReindexNoConfig(t *testing.T) {
	setConfigPath(t, "/nonexistent/config.yaml")
	setReindexTemplate(t, "")

	cmd := &cobra.Command{}
	err := runReindex(cmd, nil)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestRunReindexPreservesExisting(t *testing.T) {
	cfgDir, _ := setupReindexEnv(t)
	setReindexTemplate(t, "")
	setVerbose(t, false)

	// Pre-populate with an entry whose path doesn't exist on disk
	idxPath := filepath.Join(cfgDir, "projects.json")
	preIdx := &index.Index{
		Projects: []index.Entry{
			{Name: "OldProject", TemplateID: "other", Path: "/somewhere/else"},
		},
	}
	data, _ := json.MarshalIndent(preIdx, "", "  ")
	os.WriteFile(idxPath, data, 0644)

	cmd := &cobra.Command{}
	err := runReindex(cmd, nil)
	if err != nil {
		t.Fatalf("runReindex() error: %v", err)
	}

	idx, _ := index.Load(idxPath)
	// Should have 1 preserved + 3 new = 4
	if len(idx.Projects) != 4 {
		t.Errorf("indexed %d projects, want 4 (1 preserved + 3 new)", len(idx.Projects))
	}

	// Verify old entry is still there
	found := false
	for _, e := range idx.Projects {
		if e.Name == "OldProject" {
			found = true
			break
		}
	}
	if !found {
		t.Error("pre-existing entry was not preserved")
	}
}
