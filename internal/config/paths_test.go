package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultPath(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error: %v", err)
	}

	if path == "" {
		t.Fatal("DefaultPath() returned empty string")
	}

	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		expected := filepath.Join(home, ".prjct", "config.yaml")
		if path != expected {
			t.Errorf("DefaultPath() = %q, want %q", path, expected)
		}
	default:
		expected := filepath.Join(home, ".config", "prjct", "config.yaml")
		if path != expected {
			t.Errorf("DefaultPath() = %q, want %q", path, expected)
		}
	}
}

func TestDefaultPathContainsConfigYaml(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error: %v", err)
	}

	if filepath.Base(path) != "config.yaml" {
		t.Errorf("DefaultPath() base = %q, want %q", filepath.Base(path), "config.yaml")
	}
}
