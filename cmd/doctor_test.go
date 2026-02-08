package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunDoctorValidConfig(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Base path exists (it's a t.TempDir()), so all checks should pass.
	cmd := &cobra.Command{}
	err := runDoctor(cmd, nil)
	if err != nil {
		t.Fatalf("runDoctor() error: %v", err)
	}
}

func TestRunDoctorMissingConfig(t *testing.T) {
	setConfigPath(t, "/nonexistent/config.yaml")

	cmd := &cobra.Command{}
	err := runDoctor(cmd, nil)
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

func TestRunDoctorInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(cfgPath, []byte(":::invalid:::"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runDoctor(cmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitConfigInvalid {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitConfigInvalid)
	}
}

func TestRunDoctorEmptyTemplates(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(cfgPath, []byte("templates: []\n"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runDoctor(cmd, nil)
	if err == nil {
		t.Fatal("expected error for empty templates")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitConfigInvalid {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitConfigInvalid)
	}
}

func TestRunDoctorDuplicateIDs(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "dupes.yaml")
	content := `templates:
  - id: dupe
    name: "First"
    base_path: "/tmp/a"
    directories:
      - name: "src"
  - id: dupe
    name: "Second"
    base_path: "/tmp/b"
    directories:
      - name: "src"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runDoctor(cmd, nil)
	if err == nil {
		t.Fatal("expected error for duplicate IDs")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitConfigInvalid {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitConfigInvalid)
	}
}

func TestRunDoctorReservedID(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "reserved.yaml")
	content := `templates:
  - id: list
    name: "Conflicts with command"
    base_path: "/tmp/test"
    directories:
      - name: "src"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runDoctor(cmd, nil)
	if err == nil {
		t.Fatal("expected error for reserved ID")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitConfigInvalid {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitConfigInvalid)
	}
}

func TestRunDoctorBasePathWarnings(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	// Base path doesn't exist — doctor should still succeed (warnings only)
	content := fmt.Sprintf(`templates:
  - id: test
    name: "Test"
    base_path: %q
    directories:
      - name: "src"
`, filepath.Join(dir, "nonexistent_base"))

	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runDoctor(cmd, nil)
	// Should succeed — missing base paths are warnings, not errors
	if err != nil {
		t.Fatalf("runDoctor() error: %v (base path warnings should not fail)", err)
	}
}

func TestRunDoctorDefaultPath(t *testing.T) {
	// With empty configPath, doctor uses config.DefaultPath()
	// The default config probably doesn't exist, so this should fail with ConfigNotFound
	setConfigPath(t, "")

	cmd := &cobra.Command{}
	err := runDoctor(cmd, nil)
	// We don't assert the specific error because the default config might or might not exist.
	// Just ensure it doesn't panic.
	_ = err
}

func TestPrintCheck(t *testing.T) {
	// printCheck should not panic for any status
	statuses := []string{"OK", "WARN", "FAIL"}
	for _, s := range statuses {
		printCheck(s, "test message")
	}
}
