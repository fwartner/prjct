package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfigYAMLIsValidYAML(t *testing.T) {
	content := DefaultConfigYAML()
	if content == "" {
		t.Fatal("DefaultConfigYAML() returned empty string")
	}

	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		t.Fatalf("DefaultConfigYAML() produces invalid YAML: %v", err)
	}
}

func TestDefaultConfigYAMLHasTemplates(t *testing.T) {
	content := DefaultConfigYAML()

	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(cfg.Templates) < 3 {
		t.Errorf("expected at least 3 templates, got %d", len(cfg.Templates))
	}

	// Check expected template IDs
	ids := make(map[string]bool)
	for _, tmpl := range cfg.Templates {
		ids[tmpl.ID] = true
	}

	for _, expected := range []string{"video", "photo", "dev"} {
		if !ids[expected] {
			t.Errorf("expected template ID %q not found", expected)
		}
	}
}

func TestDefaultConfigYAMLPassesValidation(t *testing.T) {
	content := DefaultConfigYAML()

	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	errs := cfg.Validate()
	if len(errs) > 0 {
		t.Errorf("default config has validation errors: %v", errs)
	}
}

func TestDefaultConfigYAMLHasDirectories(t *testing.T) {
	content := DefaultConfigYAML()

	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	for _, tmpl := range cfg.Templates {
		if len(tmpl.Directories) == 0 {
			t.Errorf("template %q has no directories", tmpl.ID)
		}
	}
}

func TestDefaultConfigYAMLHasBasePaths(t *testing.T) {
	content := DefaultConfigYAML()

	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	for _, tmpl := range cfg.Templates {
		if tmpl.BasePath == "" {
			t.Errorf("template %q has empty base_path", tmpl.ID)
		}
		if !strings.HasPrefix(tmpl.BasePath, "~") && !strings.HasPrefix(tmpl.BasePath, "/") {
			t.Errorf("template %q base_path %q should start with ~ or /", tmpl.ID, tmpl.BasePath)
		}
	}
}
