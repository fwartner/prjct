package journal

import (
	"path/filepath"
	"testing"
	"time"
)

func TestAppendAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")

	rec := Record{
		Timestamp: time.Now(),
		Operation: OpCreate,
		Details: map[string]string{
			"path": "/tmp/test",
			"name": "test",
		},
	}

	if err := Append(path, rec); err != nil {
		t.Fatalf("Append: %v", err)
	}

	j, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(j.Records) != 1 {
		t.Fatalf("got %d records, want 1", len(j.Records))
	}
	if j.Records[0].Operation != OpCreate {
		t.Errorf("operation = %q, want %q", j.Records[0].Operation, OpCreate)
	}
	if j.Records[0].Details["path"] != "/tmp/test" {
		t.Errorf("path = %q, want /tmp/test", j.Records[0].Details["path"])
	}
}

func TestLoadNonExistent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nope.json")

	j, err := Load(path)
	if err != nil {
		t.Fatalf("Load non-existent: %v", err)
	}
	if len(j.Records) != 0 {
		t.Fatalf("got %d records, want 0", len(j.Records))
	}
}

func TestLast(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")

	// Empty journal
	rec, err := Last(path)
	if err != nil {
		t.Fatalf("Last empty: %v", err)
	}
	if rec != nil {
		t.Fatal("expected nil for empty journal")
	}

	// Add records
	_ = Append(path, Record{Timestamp: time.Now(), Operation: OpCreate, Details: map[string]string{"n": "1"}})
	_ = Append(path, Record{Timestamp: time.Now(), Operation: OpRename, Details: map[string]string{"n": "2"}})

	rec, err = Last(path)
	if err != nil {
		t.Fatalf("Last: %v", err)
	}
	if rec == nil {
		t.Fatal("expected non-nil record")
	}
	if rec.Operation != OpRename {
		t.Errorf("operation = %q, want %q", rec.Operation, OpRename)
	}
}

func TestRemoveLast(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")

	_ = Append(path, Record{Timestamp: time.Now(), Operation: OpCreate, Details: map[string]string{}})
	_ = Append(path, Record{Timestamp: time.Now(), Operation: OpRename, Details: map[string]string{}})

	if err := RemoveLast(path); err != nil {
		t.Fatalf("RemoveLast: %v", err)
	}

	j, _ := Load(path)
	if len(j.Records) != 1 {
		t.Fatalf("got %d records, want 1", len(j.Records))
	}
	if j.Records[0].Operation != OpCreate {
		t.Errorf("operation = %q, want %q", j.Records[0].Operation, OpCreate)
	}
}

func TestMaxRecords(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")

	for i := 0; i < 110; i++ {
		_ = Append(path, Record{
			Timestamp: time.Now(),
			Operation: OpCreate,
			Details:   map[string]string{"i": string(rune('A' + i%26))},
		})
	}

	j, _ := Load(path)
	if len(j.Records) > 100 {
		t.Errorf("got %d records, want <=100", len(j.Records))
	}
}
