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

func TestRunRename(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "OldName")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	entries := []index.Entry{
		{Name: "OldName", TemplateID: "video", Path: projectDir, CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))

	cmd := &cobra.Command{}
	err := runRename(cmd, []string{"OldName", "NewName"})
	if err != nil {
		t.Fatalf("runRename() error: %v", err)
	}

	newPath := filepath.Join(dir, "NewName")
	if _, statErr := os.Stat(newPath); os.IsNotExist(statErr) {
		t.Error("renamed directory does not exist")
	}
	if _, statErr := os.Stat(projectDir); !os.IsNotExist(statErr) {
		t.Error("old directory should not exist after rename")
	}
}

func TestRunRenameNoMatch(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "alpha", TemplateID: "dev", Path: "/a", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))

	cmd := &cobra.Command{}
	err := runRename(cmd, []string{"zzzzz", "NewName"})
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

func TestRunRenameInvalidNewName(t *testing.T) {
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

	cmd := &cobra.Command{}
	err := runRename(cmd, []string{"proj", "CON"})
	if err == nil {
		t.Fatal("expected error for invalid new name")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitInvalidName {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitInvalidName)
	}
}

func TestRunRenameTargetExists(t *testing.T) {
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "proj")
	existingDir := filepath.Join(dir, "existing")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(existingDir, 0755); err != nil {
		t.Fatal(err)
	}
	entries := []index.Entry{
		{Name: "proj", TemplateID: "dev", Path: projectDir, CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))

	cmd := &cobra.Command{}
	err := runRename(cmd, []string{"proj", "existing"})
	if err == nil {
		t.Fatal("expected error when target exists")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitProjectExists {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitProjectExists)
	}
}
