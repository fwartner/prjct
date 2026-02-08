package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	content := `templates:
  - id: test
    name: "Test Template"
    base_path: "/tmp/test"
    directories:
      - name: "src"
      - name: "docs"
`
	path := writeTemp(t, content)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(cfg.Templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(cfg.Templates))
	}

	tmpl := cfg.Templates[0]
	if tmpl.ID != "test" {
		t.Errorf("ID = %q, want %q", tmpl.ID, "test")
	}
	if tmpl.Name != "Test Template" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "Test Template")
	}
	if tmpl.BasePath != "/tmp/test" {
		t.Errorf("BasePath = %q, want %q", tmpl.BasePath, "/tmp/test")
	}
	if len(tmpl.Directories) != 2 {
		t.Errorf("expected 2 directories, got %d", len(tmpl.Directories))
	}
}

func TestLoadNestedDirectories(t *testing.T) {
	content := `templates:
  - id: nested
    name: "Nested"
    base_path: "/tmp"
    directories:
      - name: "parent"
        children:
          - name: "child"
            children:
              - name: "grandchild"
`
	path := writeTemp(t, content)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	dirs := cfg.Templates[0].Directories
	if len(dirs) != 1 {
		t.Fatalf("expected 1 top-level dir, got %d", len(dirs))
	}
	if dirs[0].Name != "parent" {
		t.Errorf("top dir name = %q, want %q", dirs[0].Name, "parent")
	}
	if len(dirs[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(dirs[0].Children))
	}
	if dirs[0].Children[0].Name != "child" {
		t.Errorf("child name = %q, want %q", dirs[0].Children[0].Name, "child")
	}
	if len(dirs[0].Children[0].Children) != 1 {
		t.Fatalf("expected 1 grandchild, got %d", len(dirs[0].Children[0].Children))
	}
	if dirs[0].Children[0].Children[0].Name != "grandchild" {
		t.Errorf("grandchild name = %q, want %q", dirs[0].Children[0].Children[0].Name, "grandchild")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	content := `templates:
  - id: "unclosed string
    name: broken
`
	path := writeTemp(t, content)

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() expected error for invalid YAML, got nil")
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("Load() expected error for nonexistent file, got nil")
	}
}

func TestLoadMultipleTemplates(t *testing.T) {
	content := `templates:
  - id: video
    name: "Video"
    base_path: "/tmp/video"
    directories:
      - name: "footage"
  - id: photo
    name: "Photo"
    base_path: "/tmp/photo"
    directories:
      - name: "raw"
  - id: dev
    name: "Dev"
    base_path: "/tmp/dev"
    directories:
      - name: "src"
`
	path := writeTemp(t, content)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(cfg.Templates) != 3 {
		t.Fatalf("expected 3 templates, got %d", len(cfg.Templates))
	}
}

func TestValidateValid(t *testing.T) {
	cfg := &Config{
		Templates: []Template{
			{
				ID:       "test",
				Name:     "Test",
				BasePath: "/tmp",
				Directories: []Directory{
					{Name: "src"},
				},
			},
		},
	}

	errs := cfg.Validate()
	if len(errs) > 0 {
		t.Errorf("Validate() returned %d errors for valid config: %v", len(errs), errs)
	}
}

func TestValidateNoTemplates(t *testing.T) {
	cfg := &Config{}

	errs := cfg.Validate()
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for empty templates")
	}

	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "no templates defined") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'no templates defined' error")
	}
}

func TestValidateMissingID(t *testing.T) {
	cfg := &Config{
		Templates: []Template{
			{Name: "Test", BasePath: "/tmp", Directories: []Directory{{Name: "src"}}},
		},
	}

	errs := cfg.Validate()
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for missing ID")
	}

	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "id is required") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'id is required' error")
	}
}

func TestValidateMissingName(t *testing.T) {
	cfg := &Config{
		Templates: []Template{
			{ID: "test", BasePath: "/tmp", Directories: []Directory{{Name: "src"}}},
		},
	}

	errs := cfg.Validate()
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for missing name")
	}
}

func TestValidateMissingBasePath(t *testing.T) {
	cfg := &Config{
		Templates: []Template{
			{ID: "test", Name: "Test", Directories: []Directory{{Name: "src"}}},
		},
	}

	errs := cfg.Validate()
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for missing base_path")
	}
}

func TestValidateNoDirectories(t *testing.T) {
	cfg := &Config{
		Templates: []Template{
			{ID: "test", Name: "Test", BasePath: "/tmp"},
		},
	}

	errs := cfg.Validate()
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for missing directories")
	}
}

func TestValidateDuplicateIDs(t *testing.T) {
	cfg := &Config{
		Templates: []Template{
			{ID: "test", Name: "Test 1", BasePath: "/tmp", Directories: []Directory{{Name: "a"}}},
			{ID: "test", Name: "Test 2", BasePath: "/tmp", Directories: []Directory{{Name: "b"}}},
		},
	}

	errs := cfg.Validate()
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for duplicate IDs")
	}

	found := false
	for _, e := range errs {
		if strings.Contains(e.Message, "duplicate") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'duplicate' error")
	}
}

func TestValidateReservedIDs(t *testing.T) {
	reserved := []string{"list", "config", "doctor", "help", "install"}
	for _, id := range reserved {
		cfg := &Config{
			Templates: []Template{
				{ID: id, Name: "Test", BasePath: "/tmp", Directories: []Directory{{Name: "src"}}},
			},
		}

		errs := cfg.Validate()
		if len(errs) == 0 {
			t.Errorf("Validate() expected error for reserved ID %q", id)
			continue
		}

		found := false
		for _, e := range errs {
			if strings.Contains(e.Message, "conflicts with a built-in command") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected 'conflicts with a built-in command' error for ID %q", id)
		}
	}
}

func TestValidateEmptyDirectoryName(t *testing.T) {
	cfg := &Config{
		Templates: []Template{
			{
				ID:       "test",
				Name:     "Test",
				BasePath: "/tmp",
				Directories: []Directory{
					{Name: "valid"},
					{Name: ""},
				},
			},
		},
	}

	errs := cfg.Validate()
	if len(errs) == 0 {
		t.Fatal("Validate() expected errors for empty directory name")
	}
}

func TestFindTemplate(t *testing.T) {
	cfg := &Config{
		Templates: []Template{
			{ID: "video", Name: "Video"},
			{ID: "photo", Name: "Photo"},
			{ID: "dev", Name: "Dev"},
		},
	}

	tests := []struct {
		id       string
		wantName string
		wantNil  bool
	}{
		{"video", "Video", false},
		{"photo", "Photo", false},
		{"dev", "Dev", false},
		{"nonexistent", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			tmpl := cfg.FindTemplate(tt.id)
			if tt.wantNil {
				if tmpl != nil {
					t.Errorf("FindTemplate(%q) = %v, want nil", tt.id, tmpl)
				}
			} else {
				if tmpl == nil {
					t.Fatalf("FindTemplate(%q) = nil, want template", tt.id)
				}
				if tmpl.Name != tt.wantName {
					t.Errorf("FindTemplate(%q).Name = %q, want %q", tt.id, tmpl.Name, tt.wantName)
				}
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"absolute path", "/tmp/test", "/tmp/test", false},
		{"tilde only", "~", home, false},
		{"tilde with subdir", "~/Projects", filepath.Join(home, "Projects"), false},
		{"tilde with nested", "~/Projects/Video", filepath.Join(home, "Projects", "Video"), false},
		{"relative path", "relative/path", "relative/path", false},
		{"empty path", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandPath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandPath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidationErrorString(t *testing.T) {
	e := ValidationError{Field: "templates[0].id", Message: "id is required"}
	want := "templates[0].id: id is required"
	if e.Error() != want {
		t.Errorf("Error() = %q, want %q", e.Error(), want)
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	return path
}
