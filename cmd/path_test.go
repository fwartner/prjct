package cmd

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

func setPathTemplate(t *testing.T, val string) {
	t.Helper()
	old := pathTemplate
	pathTemplate = val
	t.Cleanup(func() { pathTemplate = old })
}

func TestRunPath(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "MyProject", TemplateID: "video", Path: "/projects/MyProject", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setPathTemplate(t, "")

	cmd := &cobra.Command{}
	err := runPath(cmd, []string{"MyProject"})
	if err != nil {
		t.Fatalf("runPath() error: %v", err)
	}
}

func TestRunPathNoMatch(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "alpha", TemplateID: "dev", Path: "/a", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setPathTemplate(t, "")

	cmd := &cobra.Command{}
	err := runPath(cmd, []string{"zzzzz"})
	if err == nil {
		t.Fatal("expected error for no match")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitGeneral {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitGeneral)
	}
}

func TestRunPathWithTemplateFilter(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "proj", TemplateID: "video", Path: "/a", CreatedAt: time.Now()},
		{Name: "proj", TemplateID: "dev", Path: "/b", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setPathTemplate(t, "dev")

	cmd := &cobra.Command{}
	err := runPath(cmd, []string{"proj"})
	if err != nil {
		t.Fatalf("runPath() error: %v", err)
	}
}
