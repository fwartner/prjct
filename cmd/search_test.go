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

// writeTestIndex creates a test index file with sample entries and returns its path.
func writeTestIndex(t *testing.T, dir string, entries []index.Entry) string {
	t.Helper()
	path := filepath.Join(dir, "projects.json")
	idx := &index.Index{Projects: entries}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// setSearchTemplate sets the package-level searchTemplate for the duration of the test.
func setSearchTemplate(t *testing.T, val string) {
	t.Helper()
	old := searchTemplate
	searchTemplate = val
	t.Cleanup(func() { searchTemplate = old })
}

func TestRunSearchAllProjects(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "Video Project", TemplateID: "video", Path: "/a", CreatedAt: time.Now()},
		{Name: "Dev API", TemplateID: "dev", Path: "/b", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setSearchTemplate(t, "")

	cmd := &cobra.Command{}
	err := runSearch(cmd, []string{})
	if err != nil {
		t.Fatalf("runSearch() error: %v", err)
	}
}

func TestRunSearchWithQuery(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "Client Commercial", TemplateID: "video", Path: "/a", CreatedAt: time.Now()},
		{Name: "Product Shoot", TemplateID: "photo", Path: "/b", CreatedAt: time.Now()},
		{Name: "api-gateway", TemplateID: "dev", Path: "/c", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setSearchTemplate(t, "")

	cmd := &cobra.Command{}
	err := runSearch(cmd, []string{"commercial"})
	if err != nil {
		t.Fatalf("runSearch() error: %v", err)
	}
}

func TestRunSearchNoResults(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "alpha", TemplateID: "dev", Path: "/a", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setSearchTemplate(t, "")

	cmd := &cobra.Command{}
	err := runSearch(cmd, []string{"zzzzz"})
	if err != nil {
		t.Fatalf("runSearch() error: %v", err)
	}
}

func TestRunSearchWithTemplateFilter(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "a", TemplateID: "video", Path: "/a", CreatedAt: time.Now()},
		{Name: "b", TemplateID: "dev", Path: "/b", CreatedAt: time.Now()},
		{Name: "c", TemplateID: "video", Path: "/c", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setSearchTemplate(t, "video")

	cmd := &cobra.Command{}
	err := runSearch(cmd, []string{})
	if err != nil {
		t.Fatalf("runSearch() error: %v", err)
	}
}

func TestRunSearchMissingIndex(t *testing.T) {
	dir := t.TempDir()
	// No index file — should return empty, not error
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setSearchTemplate(t, "")

	cmd := &cobra.Command{}
	err := runSearch(cmd, []string{})
	if err != nil {
		t.Fatalf("runSearch() with missing index error: %v", err)
	}
}

func TestRunSearchCorruptIndex(t *testing.T) {
	dir := t.TempDir()
	idxPath := filepath.Join(dir, "projects.json")
	if err := os.WriteFile(idxPath, []byte("not json!!!"), 0644); err != nil {
		t.Fatal(err)
	}
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setSearchTemplate(t, "")

	cmd := &cobra.Command{}
	err := runSearch(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for corrupt index")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitGeneral {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitGeneral)
	}
}

func TestResolveIndexPathWithConfig(t *testing.T) {
	setConfigPath(t, "/some/dir/config.yaml")
	p, err := resolveIndexPath()
	if err != nil {
		t.Fatalf("resolveIndexPath() error: %v", err)
	}
	if filepath.Base(p) != "projects.json" {
		t.Errorf("resolveIndexPath() = %q, want projects.json filename", p)
	}
	if filepath.Dir(p) != "/some/dir" {
		t.Errorf("resolveIndexPath() dir = %q, want /some/dir", filepath.Dir(p))
	}
}

func TestResolveIndexPathDefault(t *testing.T) {
	setConfigPath(t, "")
	p, err := resolveIndexPath()
	if err != nil {
		t.Fatalf("resolveIndexPath() error: %v", err)
	}
	if filepath.Base(p) != "projects.json" {
		t.Errorf("resolveIndexPath() = %q, want projects.json filename", p)
	}
}
