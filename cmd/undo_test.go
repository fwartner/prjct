package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fwartner/prjct/internal/journal"
	"github.com/spf13/cobra"
)

func TestRunUndoEmpty(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	cmd := &cobra.Command{}
	err := runUndo(cmd, []string{})
	if err != nil {
		t.Fatalf("runUndo() empty: %v", err)
	}
}

func TestRunUndoCreate(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Create a directory to be undone
	projectDir := filepath.Join(base, "UndoMe")
	_ = os.MkdirAll(projectDir, 0755)

	// Write journal entry
	jPath := filepath.Join(filepath.Dir(cfgPath), "journal.json")
	_ = journal.Append(jPath, journal.Record{
		Timestamp: time.Now(),
		Operation: journal.OpCreate,
		Details: map[string]string{
			"path":     projectDir,
			"template": "test",
			"name":     "UndoMe",
		},
	})

	cmd := &cobra.Command{}
	err := runUndo(cmd, []string{})
	if err != nil {
		t.Fatalf("runUndo() error: %v", err)
	}

	// Directory should be removed
	if _, statErr := os.Stat(projectDir); !os.IsNotExist(statErr) {
		t.Error("expected project directory to be removed")
	}
}

func TestRunUndoRename(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	// Create the "new" directory (post-rename state)
	newDir := filepath.Join(base, "NewName")
	_ = os.MkdirAll(newDir, 0755)

	oldDir := filepath.Join(base, "OldName")

	// Write journal entry
	jPath := filepath.Join(filepath.Dir(cfgPath), "journal.json")
	_ = journal.Append(jPath, journal.Record{
		Timestamp: time.Now(),
		Operation: journal.OpRename,
		Details: map[string]string{
			"old_path": oldDir,
			"new_path": newDir,
			"old_name": "OldName",
			"new_name": "NewName",
		},
	})

	cmd := &cobra.Command{}
	err := runUndo(cmd, []string{})
	if err != nil {
		t.Fatalf("runUndo() error: %v", err)
	}

	if _, statErr := os.Stat(oldDir); os.IsNotExist(statErr) {
		t.Error("old directory should be restored")
	}
	if _, statErr := os.Stat(newDir); !os.IsNotExist(statErr) {
		t.Error("new directory should be gone after undo")
	}
}

func TestResolveJournalPath(t *testing.T) {
	base := t.TempDir()
	cfgPath := filepath.Join(base, "config.yaml")
	setConfigPath(t, cfgPath)

	got, err := resolveJournalPath()
	if err != nil {
		t.Fatalf("resolveJournalPath() error: %v", err)
	}
	want := filepath.Join(base, "journal.json")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
