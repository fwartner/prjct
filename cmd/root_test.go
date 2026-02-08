package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fwartner/prjct/internal/config"
	"github.com/fwartner/prjct/internal/project"
	"github.com/spf13/cobra"
)

// --- Test helpers (shared across all cmd test files) ---

// writeTestConfig writes a valid config file to a temp dir and returns its path.
func writeTestConfig(t *testing.T, basePath string) string {
	t.Helper()
	dir := t.TempDir()
	return writeTestConfigAt(t, filepath.Join(dir, "config.yaml"), basePath)
}

// writeTestConfigAt writes a valid config file at a specific path and returns it.
func writeTestConfigAt(t *testing.T, path, basePath string) string {
	t.Helper()
	content := fmt.Sprintf(`templates:
  - id: test
    name: "Test Template"
    base_path: %q
    directories:
      - name: "src"
      - name: "docs"
  - id: photo
    name: "Photography"
    base_path: %q
    directories:
      - name: "RAW"
      - name: "Edits"
`, basePath, basePath)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("writeTestConfigAt mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeTestConfigAt: %v", err)
	}
	return path
}

// setConfigPath sets the package-level configPath for the duration of the test.
func setConfigPath(t *testing.T, path string) {
	t.Helper()
	old := configPath
	configPath = path
	t.Cleanup(func() { configPath = old })
}

// setForceInstall sets the package-level forceInstall for the duration of the test.
func setForceInstall(t *testing.T, val bool) {
	t.Helper()
	old := forceInstall
	forceInstall = val
	t.Cleanup(func() { forceInstall = old })
}

// setVerbose sets the package-level verbose for the duration of the test.
func setVerbose(t *testing.T, val bool) {
	t.Helper()
	old := verbose
	verbose = val
	t.Cleanup(func() { verbose = old })
}

// withStdin replaces os.Stdin with a pipe containing the given input.
func withStdin(t *testing.T, input string) {
	t.Helper()
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = oldStdin
		r.Close()
	})
	go func() {
		defer w.Close()
		_, _ = w.WriteString(input)
	}()
}

// --- ExitError tests ---

func TestExitErrorError(t *testing.T) {
	e := &ExitError{Code: ExitGeneral, Message: "something broke"}
	if e.Error() != "something broke" {
		t.Errorf("Error() = %q, want %q", e.Error(), "something broke")
	}
}

func TestExitErrorUnwrap(t *testing.T) {
	inner := errors.New("root cause")
	e := &ExitError{Code: ExitGeneral, Message: "wrapped", Err: inner}
	if !errors.Is(e, inner) {
		t.Error("Unwrap() did not return inner error")
	}
}

func TestExitErrorUnwrapNil(t *testing.T) {
	e := &ExitError{Code: ExitGeneral, Message: "no inner"}
	if e.Unwrap() != nil {
		t.Error("Unwrap() should return nil when no inner error")
	}
}

func TestExitErrorCodes(t *testing.T) {
	codes := map[string]int{
		"OK":               ExitOK,
		"General":          ExitGeneral,
		"ConfigNotFound":   ExitConfigNotFound,
		"ConfigInvalid":    ExitConfigInvalid,
		"TemplateNotFound": ExitTemplateNotFound,
		"ProjectExists":    ExitProjectExists,
		"Permission":       ExitPermission,
		"CreateFailed":     ExitCreateFailed,
		"InvalidName":      ExitInvalidName,
		"UserCancelled":    ExitUserCancelled,
	}
	for name, code := range codes {
		if code < 0 || code > 9 {
			t.Errorf("exit code %s = %d, want 0-9", name, code)
		}
	}
}

// --- mapCreateError tests ---

func TestMapCreateErrorProjectExists(t *testing.T) {
	err := mapCreateError(fmt.Errorf("failed: %w", project.ErrProjectExists))
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitProjectExists {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitProjectExists)
	}
}

func TestMapCreateErrorPermission(t *testing.T) {
	err := mapCreateError(fmt.Errorf("failed: %w", project.ErrPermission))
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitPermission {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitPermission)
	}
}

func TestMapCreateErrorGeneric(t *testing.T) {
	err := mapCreateError(errors.New("disk full"))
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitCreateFailed {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitCreateFailed)
	}
}

// --- templateIDs tests ---

func TestTemplateIDs(t *testing.T) {
	cfg := &config.Config{
		Templates: []config.Template{
			{ID: "alpha"},
			{ID: "beta"},
			{ID: "gamma"},
		},
	}
	ids := templateIDs(cfg)
	if len(ids) != 3 {
		t.Fatalf("len(ids) = %d, want 3", len(ids))
	}
	expected := []string{"alpha", "beta", "gamma"}
	for i, want := range expected {
		if ids[i] != want {
			t.Errorf("ids[%d] = %q, want %q", i, ids[i], want)
		}
	}
}

func TestTemplateIDsEmpty(t *testing.T) {
	cfg := &config.Config{}
	ids := templateIDs(cfg)
	if len(ids) != 0 {
		t.Errorf("len(ids) = %d, want 0", len(ids))
	}
}

// --- loadConfig tests ---

func TestLoadConfigNotFound(t *testing.T) {
	setConfigPath(t, "/nonexistent/path/config.yaml")
	_, err := loadConfig()
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

func TestLoadConfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte(":::invalid yaml:::"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, path)
	_, err := loadConfig()
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

func TestLoadConfigValidationFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")
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
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, path)
	_, err := loadConfig()
	if err == nil {
		t.Fatal("expected error for validation failure")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitConfigInvalid {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitConfigInvalid)
	}
}

func TestLoadConfigValid(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error: %v", err)
	}
	if len(cfg.Templates) != 2 {
		t.Errorf("templates count = %d, want 2", len(cfg.Templates))
	}
}

// --- runRoot non-interactive tests ---

func TestRunRootNonInteractive(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{"test", "MyProject"})
	if err != nil {
		t.Fatalf("runRoot() error: %v", err)
	}

	projectPath := filepath.Join(base, "MyProject")
	if _, statErr := os.Stat(projectPath); os.IsNotExist(statErr) {
		t.Error("project directory was not created")
	}
	for _, name := range []string{"src", "docs"} {
		p := filepath.Join(projectPath, name)
		if _, statErr := os.Stat(p); os.IsNotExist(statErr) {
			t.Errorf("subdirectory %q was not created", name)
		}
	}
}

func TestRunRootNonInteractiveSecondTemplate(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{"photo", "Shoot2026"})
	if err != nil {
		t.Fatalf("runRoot() error: %v", err)
	}

	projectPath := filepath.Join(base, "Shoot2026")
	for _, name := range []string{"RAW", "Edits"} {
		p := filepath.Join(projectPath, name)
		if _, statErr := os.Stat(p); os.IsNotExist(statErr) {
			t.Errorf("subdirectory %q was not created", name)
		}
	}
}

func TestRunRootVerbose(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)
	setVerbose(t, true)

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{"test", "VerboseProject"})
	if err != nil {
		t.Fatalf("runRoot() with verbose error: %v", err)
	}

	projectPath := filepath.Join(base, "VerboseProject")
	if _, statErr := os.Stat(projectPath); os.IsNotExist(statErr) {
		t.Error("project directory was not created")
	}
}

func TestRunRootTemplateNotFound(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{"nonexistent", "MyProject"})
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitTemplateNotFound {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitTemplateNotFound)
	}
	if !strings.Contains(exitErr.Message, "nonexistent") {
		t.Errorf("message should mention template name, got: %s", exitErr.Message)
	}
	if !strings.Contains(exitErr.Message, "test") {
		t.Errorf("message should list available IDs, got: %s", exitErr.Message)
	}
}

func TestRunRootOneArg(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{"test"})
	if err == nil {
		t.Fatal("expected error for 1 arg")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitGeneral {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitGeneral)
	}
}

func TestRunRootInvalidName(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{"test", "CON"})
	if err == nil {
		t.Fatal("expected error for reserved name")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitInvalidName {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitInvalidName)
	}
}

func TestRunRootEmptyName(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{"test", ""})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitInvalidName {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitInvalidName)
	}
}

func TestRunRootProjectExists(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	if err := os.Mkdir(filepath.Join(base, "Existing"), 0755); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{"test", "Existing"})
	if err == nil {
		t.Fatal("expected error for existing project")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitProjectExists {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitProjectExists)
	}
}

func TestRunRootConfigNotFound(t *testing.T) {
	setConfigPath(t, "/nonexistent/config.yaml")

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{"test", "Project"})
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

func TestRunRootNameWithSpaces(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{"test", "Client Commercial 2026"})
	if err != nil {
		t.Fatalf("runRoot() error: %v", err)
	}

	projectPath := filepath.Join(base, "Client Commercial 2026")
	if _, statErr := os.Stat(projectPath); os.IsNotExist(statErr) {
		t.Error("project with spaces was not created")
	}
}

// --- Interactive mode tests ---

func TestRunRootInteractive(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Select template 1, name "InteractiveProject"
	withStdin(t, "1\nInteractiveProject\n")

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{})
	if err != nil {
		t.Fatalf("runRoot() interactive error: %v", err)
	}

	projectPath := filepath.Join(base, "InteractiveProject")
	if _, statErr := os.Stat(projectPath); os.IsNotExist(statErr) {
		t.Error("project was not created in interactive mode")
	}
}

func TestRunRootInteractiveSecondTemplate(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Select template 2 (photo)
	withStdin(t, "2\nPhotoShoot\n")

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{})
	if err != nil {
		t.Fatalf("runRoot() error: %v", err)
	}

	for _, name := range []string{"RAW", "Edits"} {
		p := filepath.Join(base, "PhotoShoot", name)
		if _, statErr := os.Stat(p); os.IsNotExist(statErr) {
			t.Errorf("subdirectory %q was not created", name)
		}
	}
}

func TestRunRootInteractiveCancelEmpty(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Empty selection = cancel
	withStdin(t, "\n")

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for cancelled selection")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitUserCancelled {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitUserCancelled)
	}
}

func TestRunRootInteractiveInvalidSelection(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	withStdin(t, "abc\n")

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for invalid selection")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitGeneral {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitGeneral)
	}
}

func TestRunRootInteractiveOutOfRange(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Only 2 templates, select 5
	withStdin(t, "5\n")

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for out-of-range selection")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitGeneral {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitGeneral)
	}
}

func TestRunRootInteractiveEmptyName(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Valid selection, empty name
	withStdin(t, "1\n\n")

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for empty project name")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitUserCancelled {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitUserCancelled)
	}
}

func TestRunRootInteractiveEOF(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// EOF immediately
	withStdin(t, "")

	cmd := &cobra.Command{}
	err := runRoot(cmd, []string{})
	if err == nil {
		t.Fatal("expected error for EOF")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitUserCancelled {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitUserCancelled)
	}
}
