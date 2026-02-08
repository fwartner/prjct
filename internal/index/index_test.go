package index

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIndexPath(t *testing.T) {
	p, err := IndexPath()
	if err != nil {
		t.Fatalf("IndexPath() error: %v", err)
	}
	if filepath.Base(p) != "projects.json" {
		t.Errorf("IndexPath() = %q, want filename projects.json", p)
	}
}

func TestLoadMissingFile(t *testing.T) {
	idx, err := Load("/nonexistent/projects.json")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(idx.Projects) != 0 {
		t.Errorf("Load() missing file returned %d projects, want 0", len(idx.Projects))
	}
}

func TestLoadValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.json")

	idx := &Index{
		Projects: []Entry{
			{Name: "test", TemplateID: "dev", Path: "/tmp/test", CreatedAt: time.Now()},
		},
	}
	data, _ := json.MarshalIndent(idx, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(loaded.Projects) != 1 {
		t.Fatalf("Load() returned %d projects, want 1", len(loaded.Projects))
	}
	if loaded.Projects[0].Name != "test" {
		t.Errorf("Name = %q, want %q", loaded.Projects[0].Name, "test")
	}
}

func TestLoadCorruptJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.json")
	if err := os.WriteFile(path, []byte("not json!!!"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() expected error for corrupt JSON")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.json")

	idx := &Index{
		Projects: []Entry{
			{Name: "alpha", TemplateID: "video", TemplateName: "Video", Path: "/a", CreatedAt: time.Now()},
			{Name: "beta", TemplateID: "dev", TemplateName: "Dev", Path: "/b", CreatedAt: time.Now()},
		},
	}

	if err := Save(path, idx); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(loaded.Projects) != 2 {
		t.Errorf("round-trip: got %d projects, want 2", len(loaded.Projects))
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "projects.json")

	idx := &Index{Projects: []Entry{{Name: "test", Path: "/test"}}}
	if err := Save(path, idx); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Save() did not create nested directory")
	}
}

func TestAddEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.json")

	entry := Entry{
		Name:         "MyProject",
		TemplateID:   "video",
		TemplateName: "Video Production",
		Path:         "/home/user/Projects/Video/MyProject",
		CreatedAt:    time.Now(),
	}

	if err := Add(path, entry); err != nil {
		t.Fatalf("Add() error: %v", err)
	}

	idx, _ := Load(path)
	if len(idx.Projects) != 1 {
		t.Fatalf("Add() resulted in %d projects, want 1", len(idx.Projects))
	}
	if idx.Projects[0].Name != "MyProject" {
		t.Errorf("Name = %q, want %q", idx.Projects[0].Name, "MyProject")
	}
}

func TestAddDuplicateSkipped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.json")

	entry := Entry{Name: "Proj", Path: "/same/path", CreatedAt: time.Now()}

	if err := Add(path, entry); err != nil {
		t.Fatal(err)
	}
	// Add again with same Path
	if err := Add(path, Entry{Name: "Different", Path: "/same/path", CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	idx, _ := Load(path)
	if len(idx.Projects) != 1 {
		t.Errorf("duplicate Add() resulted in %d projects, want 1", len(idx.Projects))
	}
	if idx.Projects[0].Name != "Proj" {
		t.Errorf("original entry overwritten: Name = %q, want %q", idx.Projects[0].Name, "Proj")
	}
}

func TestAddMultiple(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.json")

	for i, name := range []string{"a", "b", "c"} {
		entry := Entry{Name: name, Path: filepath.Join("/projects", name), CreatedAt: time.Now().Add(time.Duration(i) * time.Hour)}
		if err := Add(path, entry); err != nil {
			t.Fatalf("Add(%q) error: %v", name, err)
		}
	}

	idx, _ := Load(path)
	if len(idx.Projects) != 3 {
		t.Errorf("got %d projects, want 3", len(idx.Projects))
	}
}

func TestRemoveEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.json")

	_ = Add(path, Entry{Name: "keep", Path: "/keep"})
	_ = Add(path, Entry{Name: "remove", Path: "/remove"})

	if err := Remove(path, "/remove"); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	idx, _ := Load(path)
	if len(idx.Projects) != 1 {
		t.Fatalf("Remove() resulted in %d projects, want 1", len(idx.Projects))
	}
	if idx.Projects[0].Name != "keep" {
		t.Errorf("wrong entry removed: Name = %q, want %q", idx.Projects[0].Name, "keep")
	}
}

func TestRemoveNonexistent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.json")

	_ = Add(path, Entry{Name: "only", Path: "/only"})

	if err := Remove(path, "/does-not-exist"); err != nil {
		t.Fatalf("Remove() error: %v", err)
	}

	idx, _ := Load(path)
	if len(idx.Projects) != 1 {
		t.Errorf("Remove() nonexistent changed count: got %d, want 1", len(idx.Projects))
	}
}

func TestSearchByName(t *testing.T) {
	idx := &Index{
		Projects: []Entry{
			{Name: "Client Commercial", TemplateID: "video"},
			{Name: "Product Shoot", TemplateID: "photo"},
			{Name: "api-gateway", TemplateID: "dev"},
		},
	}

	results := Search(idx, "commercial")
	if len(results) != 1 {
		t.Fatalf("Search(commercial) returned %d, want 1", len(results))
	}
	if results[0].Name != "Client Commercial" {
		t.Errorf("Name = %q, want %q", results[0].Name, "Client Commercial")
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	idx := &Index{
		Projects: []Entry{
			{Name: "MyProject", TemplateID: "dev"},
		},
	}

	for _, q := range []string{"myproject", "MYPROJECT", "MyProject", "myPROJECT"} {
		results := Search(idx, q)
		if len(results) != 1 {
			t.Errorf("Search(%q) returned %d, want 1", q, len(results))
		}
	}
}

func TestSearchByTemplateID(t *testing.T) {
	idx := &Index{
		Projects: []Entry{
			{Name: "a", TemplateID: "video"},
			{Name: "b", TemplateID: "photo"},
			{Name: "c", TemplateID: "video"},
		},
	}

	results := Search(idx, "video")
	if len(results) != 2 {
		t.Errorf("Search(video) returned %d, want 2", len(results))
	}
}

func TestSearchByTemplateName(t *testing.T) {
	idx := &Index{
		Projects: []Entry{
			{Name: "proj", TemplateName: "Video Production"},
		},
	}

	results := Search(idx, "production")
	if len(results) != 1 {
		t.Errorf("Search(production) returned %d, want 1", len(results))
	}
}

func TestSearchByPath(t *testing.T) {
	idx := &Index{
		Projects: []Entry{
			{Name: "proj", Path: "/home/user/Projects/Video/proj"},
		},
	}

	results := Search(idx, "/video/")
	if len(results) != 1 {
		t.Errorf("Search(/video/) returned %d, want 1", len(results))
	}
}

func TestSearchNoMatch(t *testing.T) {
	idx := &Index{
		Projects: []Entry{
			{Name: "alpha"},
			{Name: "beta"},
		},
	}

	results := Search(idx, "zzzzz")
	if len(results) != 0 {
		t.Errorf("Search(zzzzz) returned %d, want 0", len(results))
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	idx := &Index{
		Projects: []Entry{
			{Name: "a"},
			{Name: "b"},
			{Name: "c"},
		},
	}

	results := Search(idx, "")
	if len(results) != 3 {
		t.Errorf("Search('') returned %d, want 3 (all)", len(results))
	}
}

func TestSearchEmptyIndex(t *testing.T) {
	idx := &Index{}
	results := Search(idx, "anything")
	if len(results) != 0 {
		t.Errorf("Search on empty index returned %d, want 0", len(results))
	}
}

func TestFilterByTemplate(t *testing.T) {
	entries := []Entry{
		{Name: "a", TemplateID: "video"},
		{Name: "b", TemplateID: "photo"},
		{Name: "c", TemplateID: "video"},
		{Name: "d", TemplateID: "dev"},
	}

	results := FilterByTemplate(entries, "video")
	if len(results) != 2 {
		t.Errorf("FilterByTemplate(video) returned %d, want 2", len(results))
	}
	for _, r := range results {
		if r.TemplateID != "video" {
			t.Errorf("filtered entry has TemplateID = %q, want video", r.TemplateID)
		}
	}
}

func TestFilterByTemplateNoMatch(t *testing.T) {
	entries := []Entry{
		{Name: "a", TemplateID: "video"},
	}

	results := FilterByTemplate(entries, "nonexistent")
	if len(results) != 0 {
		t.Errorf("FilterByTemplate(nonexistent) returned %d, want 0", len(results))
	}
}
