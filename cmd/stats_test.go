package cmd

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

func TestRunStats(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "a", TemplateID: "video", TemplateName: "Video", Path: "/a", CreatedAt: time.Now()},
		{Name: "b", TemplateID: "dev", TemplateName: "Dev", Path: "/b", CreatedAt: time.Now()},
		{Name: "c", TemplateID: "video", TemplateName: "Video", Path: "/c", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))

	cmd := &cobra.Command{}
	err := runStats(cmd, nil)
	if err != nil {
		t.Fatalf("runStats() error: %v", err)
	}
}

func TestRunStatsEmpty(t *testing.T) {
	dir := t.TempDir()
	writeTestIndex(t, dir, []index.Entry{})
	setConfigPath(t, filepath.Join(dir, "config.yaml"))

	cmd := &cobra.Command{}
	err := runStats(cmd, nil)
	if err != nil {
		t.Fatalf("runStats() empty error: %v", err)
	}
}

func TestRunStatsMissingIndex(t *testing.T) {
	dir := t.TempDir()
	setConfigPath(t, filepath.Join(dir, "config.yaml"))

	cmd := &cobra.Command{}
	err := runStats(cmd, nil)
	if err != nil {
		t.Fatalf("runStats() missing index error: %v", err)
	}
}
