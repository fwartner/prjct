package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

func TestRunNoteAdd(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	idx := &index.Index{
		Projects: []index.Entry{
			{Name: "NoteProject", TemplateID: "test", Path: filepath.Join(base, "NoteProject"), CreatedAt: time.Now()},
		},
	}
	data, _ := json.MarshalIndent(idx, "", "  ")
	_ = os.WriteFile(idxPath, data, 0644)

	cmd := &cobra.Command{}
	err := runNote(cmd, []string{"NoteProject", "Client: Acme Corp"})
	if err != nil {
		t.Fatalf("runNote() error: %v", err)
	}

	// Verify note was saved
	updated, _ := index.Load(idxPath)
	if len(updated.Projects[0].Notes) != 1 {
		t.Fatalf("notes count = %d, want 1", len(updated.Projects[0].Notes))
	}
	if updated.Projects[0].Notes[0] != "Client: Acme Corp" {
		t.Errorf("note = %q, want %q", updated.Projects[0].Notes[0], "Client: Acme Corp")
	}
}

func TestRunNoteView(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	idx := &index.Index{
		Projects: []index.Entry{
			{Name: "ViewNote", TemplateID: "test", Path: filepath.Join(base, "ViewNote"), CreatedAt: time.Now(), Notes: []string{"existing note"}},
		},
	}
	data, _ := json.MarshalIndent(idx, "", "  ")
	_ = os.WriteFile(idxPath, data, 0644)

	cmd := &cobra.Command{}
	err := runNote(cmd, []string{"ViewNote"})
	if err != nil {
		t.Fatalf("runNote() view error: %v", err)
	}
}

func TestRunNoteNoMatch(t *testing.T) {
	base := t.TempDir()
	cfgPath := writeTestConfig(t, base)
	setConfigPath(t, cfgPath)

	idxPath := filepath.Join(filepath.Dir(cfgPath), "projects.json")
	_ = os.WriteFile(idxPath, []byte(`{"projects":[]}`), 0644)

	cmd := &cobra.Command{}
	err := runNote(cmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
}
