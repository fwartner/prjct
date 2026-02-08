package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
)

// setEditConfig sets the package-level editConfig for the duration of the test.
func setEditConfig(t *testing.T, val bool) {
	t.Helper()
	old := editConfig
	editConfig = val
	t.Cleanup(func() { editConfig = old })
}

// setExecCommand overrides the exec function for the duration of the test.
func setExecCommand(t *testing.T, fn func(string, ...string) error) {
	t.Helper()
	old := execCommand
	execCommand = fn
	t.Cleanup(func() { execCommand = old })
}

func TestRunConfigWithExistingFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("templates: []"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)
	setEditConfig(t, false)

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err != nil {
		t.Fatalf("runConfig() error: %v", err)
	}
}

func TestRunConfigWithMissingFile(t *testing.T) {
	setConfigPath(t, "/nonexistent/path/config.yaml")
	setEditConfig(t, false)

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err != nil {
		t.Fatalf("runConfig() error: %v", err)
	}
}

func TestRunConfigCustomPath(t *testing.T) {
	dir := t.TempDir()
	customPath := filepath.Join(dir, "custom", "myconfig.yaml")
	setConfigPath(t, customPath)
	setEditConfig(t, false)

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err != nil {
		t.Fatalf("runConfig() error: %v", err)
	}
}

func TestRunConfigDefaultPath(t *testing.T) {
	setConfigPath(t, "")
	setEditConfig(t, false)

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err != nil {
		t.Fatalf("runConfig() with default path error: %v", err)
	}
}

func TestRunConfigEditOpensEditor(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("templates: []"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)
	setEditConfig(t, true)

	var calledName string
	var calledArgs []string
	setExecCommand(t, func(name string, args ...string) error {
		calledName = name
		calledArgs = args
		return nil
	})

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err != nil {
		t.Fatalf("runConfig --edit error: %v", err)
	}

	if calledName == "" {
		t.Fatal("editor was not called")
	}
	// Last arg should be the config path
	if len(calledArgs) == 0 || calledArgs[len(calledArgs)-1] != cfgPath {
		t.Errorf("editor args = %v, want last arg to be %q", calledArgs, cfgPath)
	}
}

func TestRunConfigEditMissingFile(t *testing.T) {
	setConfigPath(t, "/nonexistent/path/config.yaml")
	setEditConfig(t, true)

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err == nil {
		t.Fatal("expected error for --edit with missing file")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitConfigNotFound {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitConfigNotFound)
	}
}

func TestRunConfigEditEditorFails(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("templates: []"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)
	setEditConfig(t, true)
	setExecCommand(t, func(name string, args ...string) error {
		return errors.New("editor crashed")
	})

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err == nil {
		t.Fatal("expected error when editor fails")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitGeneral {
		t.Errorf("Code = %d, want %d", exitErr.Code, ExitGeneral)
	}
}

func TestRunConfigEditUsesConfigEditor(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := "editor: \"myeditor --wait\"\ntemplates: []\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)
	setEditConfig(t, true)

	var calledName string
	var calledArgs []string
	setExecCommand(t, func(name string, args ...string) error {
		calledName = name
		calledArgs = args
		return nil
	})

	// Clear env vars to ensure config takes priority
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err != nil {
		t.Fatalf("runConfig --edit error: %v", err)
	}

	if calledName != "myeditor" {
		t.Errorf("editor name = %q, want %q", calledName, "myeditor")
	}
	if len(calledArgs) < 2 || calledArgs[0] != "--wait" {
		t.Errorf("editor args = %v, want [--wait <path>]", calledArgs)
	}
}

func TestRunConfigEditUsesVisualEnv(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("templates: []"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)
	setEditConfig(t, true)
	t.Setenv("VISUAL", "custom-visual")
	t.Setenv("EDITOR", "should-not-use")

	var calledName string
	setExecCommand(t, func(name string, args ...string) error {
		calledName = name
		return nil
	})

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err != nil {
		t.Fatalf("runConfig --edit error: %v", err)
	}

	if calledName != "custom-visual" {
		t.Errorf("editor = %q, want %q", calledName, "custom-visual")
	}
}

func TestRunConfigEditUsesEditorEnv(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("templates: []"), 0644); err != nil {
		t.Fatal(err)
	}

	setConfigPath(t, cfgPath)
	setEditConfig(t, true)
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "custom-editor")

	var calledName string
	setExecCommand(t, func(name string, args ...string) error {
		calledName = name
		return nil
	})

	cmd := &cobra.Command{}
	err := runConfig(cmd, nil)
	if err != nil {
		t.Fatalf("runConfig --edit error: %v", err)
	}

	if calledName != "custom-editor" {
		t.Errorf("editor = %q, want %q", calledName, "custom-editor")
	}
}

func TestDefaultEditor(t *testing.T) {
	editor := defaultEditor()
	switch runtime.GOOS {
	case "darwin":
		if editor != "open" {
			t.Errorf("defaultEditor() on darwin = %q, want %q", editor, "open")
		}
	case "windows":
		if editor != "notepad" {
			t.Errorf("defaultEditor() on windows = %q, want %q", editor, "notepad")
		}
	default:
		if editor != "vi" {
			t.Errorf("defaultEditor() on %s = %q, want %q", runtime.GOOS, editor, "vi")
		}
	}
}

func TestResolveEditorPriority(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("templates: []"), 0644); err != nil {
		t.Fatal(err)
	}

	// No config editor, no env vars → platform default
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	editor := resolveEditor(cfgPath)
	expected := defaultEditor()
	if editor != expected {
		t.Errorf("resolveEditor() = %q, want default %q", editor, expected)
	}
}
