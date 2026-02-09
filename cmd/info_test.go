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

func TestRunInfo(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Create project with content
	projectDir := filepath.Join(base, "InfoProject")
	_ = os.MkdirAll(filepath.Join(projectDir, "src"), 0755)
	_ = os.WriteFile(filepath.Join(projectDir, "src", "main.go"), []byte("package main"), 0644)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	idx := &index.Index{
		Projects: []index.Entry{
			{
				Name:         "InfoProject",
				TemplateID:   "test",
				TemplateName: "Test Template",
				Path:         projectDir,
				CreatedAt:    time.Now(),
				Notes:        []string{"Test note"},
			},
		},
	}
	data, _ := json.MarshalIndent(idx, "", "  ")
	_ = os.WriteFile(idxPath, data, 0644)

	cmd := &cobra.Command{}
	err := runInfo(cmd, []string{"InfoProject"})
	if err != nil {
		t.Fatalf("runInfo() error: %v", err)
	}
}

func TestRunInfoNoMatch(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	_ = os.WriteFile(idxPath, []byte(`{"projects":[]}`), 0644)

	cmd := &cobra.Command{}
	err := runInfo(cmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
}

func TestRunInfoMissingDir(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	idx := &index.Index{
		Projects: []index.Entry{
			{Name: "Gone", TemplateID: "test", Path: "/nonexistent/gone", CreatedAt: time.Now()},
		},
	}
	data, _ := json.MarshalIndent(idx, "", "  ")
	_ = os.WriteFile(idxPath, data, 0644)

	// Should not error, just show a warning
	cmd := &cobra.Command{}
	err := runInfo(cmd, []string{"Gone"})
	if err != nil {
		t.Fatalf("runInfo() error: %v", err)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	for _, tt := range tests {
		got := formatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}
