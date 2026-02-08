package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunConfigWithExistingFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("templates: []"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err != nil {
		t.Fatalf("runConfig() error: %v", err)
	}
}

func TestRunConfigWithMissingFile(t *testing.T) {
	setConfigPath(t, "/nonexistent/path/config.yaml")

	cmd := &cobra.Command{}
	// runConfig should succeed even if the file doesn't exist
	// (it prints "not found" status, but doesn't return an error)
	err := runConfig(cmd, nil)
	if err != nil {
		t.Fatalf("runConfig() error: %v", err)
	}
}

func TestRunConfigCustomPath(t *testing.T) {
	dir := t.TempDir()
	customPath := filepath.Join(dir, "custom", "myconfig.yaml")
	setConfigPath(t, customPath)

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err != nil {
		t.Fatalf("runConfig() error: %v", err)
	}
}

func TestRunConfigDefaultPath(t *testing.T) {
	// With empty configPath, it should use config.DefaultPath()
	setConfigPath(t, "")

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err != nil {
		t.Fatalf("runConfig() with default path error: %v", err)
	}
}
