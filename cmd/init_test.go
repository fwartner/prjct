package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func setInitFlags(t *testing.T, id, name, output string) {
	t.Helper()
	oldID := initID
	oldName := initName
	oldOutput := initOutput
	initID = id
	initName = name
	initOutput = output
	t.Cleanup(func() {
		initID = oldID
		initName = oldName
		initOutput = oldOutput
	})
}

func TestRunInit(t *testing.T) {
	// Create a source directory structure
	srcDir := filepath.Join(t.TempDir(), "myproject")
	for _, d := range []string{"src", "docs", "src/components"} {
		if err := os.MkdirAll(filepath.Join(srcDir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	outFile := filepath.Join(t.TempDir(), "output.yaml")
	setInitFlags(t, "", "", outFile)

	cmd := &cobra.Command{}
	err := runInit(cmd, []string{srcDir})
	if err != nil {
		t.Fatalf("runInit() error: %v", err)
	}

	data, readErr := os.ReadFile(outFile)
	if readErr != nil {
		t.Fatalf("reading output: %v", readErr)
	}
	content := string(data)
	if !strings.Contains(content, "src") {
		t.Error("output should contain 'src' directory")
	}
	if !strings.Contains(content, "docs") {
		t.Error("output should contain 'docs' directory")
	}
}

func TestRunInitWithCustomIDAndName(t *testing.T) {
	srcDir := filepath.Join(t.TempDir(), "myproject")
	if err := os.MkdirAll(filepath.Join(srcDir, "data"), 0755); err != nil {
		t.Fatal(err)
	}

	outFile := filepath.Join(t.TempDir(), "custom.yaml")
	setInitFlags(t, "custom-id", "Custom Name", outFile)

	cmd := &cobra.Command{}
	err := runInit(cmd, []string{srcDir})
	if err != nil {
		t.Fatalf("runInit() custom error: %v", err)
	}

	data, readErr := os.ReadFile(outFile)
	if readErr != nil {
		t.Fatalf("reading output: %v", readErr)
	}
	content := string(data)
	if !strings.Contains(content, "custom-id") {
		t.Error("output should contain custom ID")
	}
	if !strings.Contains(content, "Custom Name") {
		t.Error("output should contain custom name")
	}
}

func TestRunInitStdout(t *testing.T) {
	srcDir := filepath.Join(t.TempDir(), "proj")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	setInitFlags(t, "", "", "")

	cmd := &cobra.Command{}
	err := runInit(cmd, []string{srcDir})
	if err != nil {
		t.Fatalf("runInit() stdout error: %v", err)
	}
}

func TestRunInitSkipsHiddenDirs(t *testing.T) {
	srcDir := filepath.Join(t.TempDir(), "proj")
	for _, d := range []string{"visible", ".hidden"} {
		if err := os.MkdirAll(filepath.Join(srcDir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	outFile := filepath.Join(t.TempDir(), "out.yaml")
	setInitFlags(t, "", "", outFile)

	cmd := &cobra.Command{}
	err := runInit(cmd, []string{srcDir})
	if err != nil {
		t.Fatalf("runInit() hidden error: %v", err)
	}

	data, _ := os.ReadFile(outFile)
	content := string(data)
	if strings.Contains(content, ".hidden") {
		t.Error("output should not contain hidden directory")
	}
	if !strings.Contains(content, "visible") {
		t.Error("output should contain visible directory")
	}
}

func TestRunInitNonexistentPath(t *testing.T) {
	setInitFlags(t, "", "", "")

	cmd := &cobra.Command{}
	err := runInit(cmd, []string{"/nonexistent/path"})
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
}

func TestRunInitNotADirectory(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	setInitFlags(t, "", "", "")

	cmd := &cobra.Command{}
	err := runInit(cmd, []string{tmpFile})
	if err == nil {
		t.Fatal("expected error for non-directory")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
}
