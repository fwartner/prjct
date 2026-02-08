package cmd

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

func setOpenTerminal(t *testing.T, val bool) {
	t.Helper()
	old := openTerminal
	openTerminal = val
	t.Cleanup(func() { openTerminal = old })
}

func TestRunOpen(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "MyProject", TemplateID: "video", Path: "/projects/MyProject", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setOpenTerminal(t, false)

	var calledName string
	setExecCommand(t, func(name string, args ...string) error {
		calledName = name
		return nil
	})

	cmd := &cobra.Command{}
	err := runOpen(cmd, []string{"MyProject"})
	if err != nil {
		t.Fatalf("runOpen() error: %v", err)
	}
	if calledName == "" {
		t.Fatal("expected external command to be called")
	}
}

func TestRunOpenTerminal(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "MyProject", TemplateID: "video", Path: "/projects/MyProject", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setOpenTerminal(t, true)

	var calledName string
	setExecCommand(t, func(name string, args ...string) error {
		calledName = name
		return nil
	})

	cmd := &cobra.Command{}
	err := runOpen(cmd, []string{"MyProject"})
	if err != nil {
		t.Fatalf("runOpen() terminal error: %v", err)
	}
	if calledName == "" {
		t.Fatal("expected terminal command to be called")
	}
}

func TestRunOpenNoMatch(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "alpha", TemplateID: "dev", Path: "/a", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setOpenTerminal(t, false)

	cmd := &cobra.Command{}
	err := runOpen(cmd, []string{"zzzzz"})
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

func TestRunOpenCommandFails(t *testing.T) {
	dir := t.TempDir()
	entries := []index.Entry{
		{Name: "proj", TemplateID: "dev", Path: "/projects/proj", CreatedAt: time.Now()},
	}
	writeTestIndex(t, dir, entries)
	setConfigPath(t, filepath.Join(dir, "config.yaml"))
	setOpenTerminal(t, false)

	setExecCommand(t, func(name string, args ...string) error {
		return errors.New("command failed")
	})

	cmd := &cobra.Command{}
	err := runOpen(cmd, []string{"proj"})
	if err == nil {
		t.Fatal("expected error when command fails")
	}
}
