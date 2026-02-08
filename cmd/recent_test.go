package cmd

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

func TestRunRecent(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	entries := []index.Entry{
		{Name: "old", TemplateID: "dev", Path: "/a", CreatedAt: now.Add(-24 * time.Hour)},
		{Name: "new", TemplateID: "video", Path: "/b", CreatedAt: now},
		{Name: "mid", TemplateID: "photo", Path: "/c", CreatedAt: now.Add(-1 * time.Hour)},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))

	cmd := &cobra.Command{}
	err := runRecent(cmd, []string{})
	if err != nil {
		t.Fatalf("runRecent() error: %v", err)
	}
}

func TestRunRecentWithLimit(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	entries := []index.Entry{
		{Name: "a", TemplateID: "dev", Path: "/a", CreatedAt: now},
		{Name: "b", TemplateID: "dev", Path: "/b", CreatedAt: now.Add(-1 * time.Hour)},
		{Name: "c", TemplateID: "dev", Path: "/c", CreatedAt: now.Add(-2 * time.Hour)},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))

	cmd := &cobra.Command{}
	err := runRecent(cmd, []string{"2"})
	if err != nil {
		t.Fatalf("runRecent(2) error: %v", err)
	}
}

func TestRunRecentEmpty(t *testing.T) {
	dir := t.TempDir()
	writeTestIndex(t, dir, []index.Entry{})
	setConfigPath(t, filepath.Join(dir, "config.yaml"))

	cmd := &cobra.Command{}
	err := runRecent(cmd, []string{})
	if err != nil {
		t.Fatalf("runRecent() empty error: %v", err)
	}
}

func TestRunRecentInvalidCount(t *testing.T) {
	dir := t.TempDir()
	writeTestIndex(t, dir, []index.Entry{})
	setConfigPath(t, filepath.Join(dir, "config.yaml"))

	cmd := &cobra.Command{}
	err := runRecent(cmd, []string{"abc"})
	if err == nil {
		t.Fatal("expected error for invalid count")
	}
}
