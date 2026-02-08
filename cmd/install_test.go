package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestInstallNewConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "subdir", "config.yaml")
	setConfigPath(t, cfgPath)
	setForceInstall(t, false)

	cmd := &cobra.Command{}
	err := runInstall(cmd, nil)
	if err != nil {
		t.Fatalf("runInstall() error: %v", err)
	}

	content, readErr := os.ReadFile(cfgPath)
	if readErr != nil {
		t.Fatalf("config file not created: %v", readErr)
	}
	if !strings.Contains(string(content), "templates:") {
		t.Error("config file missing templates key")
	}
}

func TestInstallForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)
	setForceInstall(t, true)

	cmd := &cobra.Command{}
	err := runInstall(cmd, nil)
	if err != nil {
		t.Fatalf("runInstall() error: %v", err)
	}

	content, readErr := os.ReadFile(cfgPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) == "old content" {
		t.Error("config was not overwritten with --force")
	}
	if !strings.Contains(string(content), "templates:") {
		t.Error("config file missing templates key after overwrite")
	}
}

func TestInstallExistingPromptYes(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)
	setForceInstall(t, false)
	withStdin(t, "y\n")

	cmd := &cobra.Command{}
	err := runInstall(cmd, nil)
	if err != nil {
		t.Fatalf("runInstall() error: %v", err)
	}

	content, readErr := os.ReadFile(cfgPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) == "old content" {
		t.Error("config was not overwritten after answering yes")
	}
}

func TestInstallExistingPromptNo(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)
	setForceInstall(t, false)
	withStdin(t, "n\n")

	cmd := &cobra.Command{}
	err := runInstall(cmd, nil)
	if err != nil {
		t.Fatalf("runInstall() error: %v", err)
	}

	content, readErr := os.ReadFile(cfgPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "old content" {
		t.Error("config should not have been overwritten after answering no")
	}
}

func TestInstallExistingPromptEOF(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)
	setForceInstall(t, false)
	withStdin(t, "")

	cmd := &cobra.Command{}
	err := runInstall(cmd, nil)
	if err == nil {
		t.Fatal("expected error for EOF on prompt")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitUserCancelled {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitUserCancelled)
	}
}

func TestInstallCreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "a", "b", "c", "config.yaml")
	setConfigPath(t, cfgPath)
	setForceInstall(t, false)

	cmd := &cobra.Command{}
	err := runInstall(cmd, nil)
	if err != nil {
		t.Fatalf("runInstall() error: %v", err)
	}

	if _, statErr := os.Stat(cfgPath); os.IsNotExist(statErr) {
		t.Error("config file was not created in nested directory")
	}
}

func TestInstallConfigIsValidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	setConfigPath(t, cfgPath)
	setForceInstall(t, false)

	cmd := &cobra.Command{}
	if err := runInstall(cmd, nil); err != nil {
		t.Fatalf("runInstall() error: %v", err)
	}

	// The installed config should be loadable and valid
	setConfigPath(t, cfgPath)
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("installed config is not valid: %v", err)
	}
	if len(cfg.Templates) == 0 {
		t.Error("installed config has no templates")
	}
}
