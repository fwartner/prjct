package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunListSuccess(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runList(cmd, nil)
	if err != nil {
		t.Fatalf("runList() error: %v", err)
	}
}

func TestRunListNoConfig(t *testing.T) {
	setConfigPath(t, "/nonexistent/config.yaml")

	cmd := &cobra.Command{}
	err := runList(cmd, nil)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitConfigNotFound {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitConfigNotFound)
	}
}

func TestRunListInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(cfgPath, []byte(":::bad:::"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runList(cmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitConfigInvalid {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitConfigInvalid)
	}
}
