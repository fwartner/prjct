package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

func setArchiveFlags(t *testing.T, del bool, output string) {
	t.Helper()
	oldDel := archiveDelete
	oldOut := archiveOutput
	archiveDelete = del
	archiveOutput = output
	t.Cleanup(func() {
		archiveDelete = oldDel
		archiveOutput = oldOut
	})
}

func TestRunArchive(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "myproject")
	if err := os.MkdirAll(filepath.Join(projectDir, "src"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "src", "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	entries := []index.Entry{
		{Name: "myproject", TemplateID: "dev", Path: projectDir, CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setArchiveFlags(t, false, "")

	cmd := &cobra.Command{}
	err := runArchive(cmd, []string{"myproject"})
	if err != nil {
		t.Fatalf("runArchive() error: %v", err)
	}

	archivePath := projectDir + ".tar.gz"
	if _, statErr := os.Stat(archivePath); os.IsNotExist(statErr) {
		t.Error("archive file was not created")
	}
	// Original should still exist
	if _, statErr := os.Stat(projectDir); os.IsNotExist(statErr) {
		t.Error("original directory should still exist without --delete")
	}
}

func TestRunArchiveWithDelete(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "todelete")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	entries := []index.Entry{
		{Name: "todelete", TemplateID: "dev", Path: projectDir, CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setArchiveFlags(t, true, "")

	cmd := &cobra.Command{}
	err := runArchive(cmd, []string{"todelete"})
	if err != nil {
		t.Fatalf("runArchive(--delete) error: %v", err)
	}

	// Original should be gone
	if _, statErr := os.Stat(projectDir); !os.IsNotExist(statErr) {
		t.Error("original directory should be deleted with --delete")
	}
}

func TestRunArchiveCustomOutput(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "proj")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	entries := []index.Entry{
		{Name: "proj", TemplateID: "dev", Path: projectDir, CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))

	customOut := filepath.Join(dir, "custom.tar.gz")
	setArchiveFlags(t, false, customOut)

	cmd := &cobra.Command{}
	err := runArchive(cmd, []string{"proj"})
	if err != nil {
		t.Fatalf("runArchive(-o) error: %v", err)
	}

	if _, statErr := os.Stat(customOut); os.IsNotExist(statErr) {
		t.Error("custom output file was not created")
	}
}

func TestRunArchiveNoMatch(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "alpha", TemplateID: "dev", Path: "/a", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setArchiveFlags(t, false, "")

	cmd := &cobra.Command{}
	err := runArchive(cmd, []string{"zzzzz"})
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

func TestRunArchiveMissingDir(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "gone", TemplateID: "dev", Path: filepath.Join(dir, "nonexistent"), CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setArchiveFlags(t, false, "")

	cmd := &cobra.Command{}
	err := runArchive(cmd, []string{"gone"})
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
}
