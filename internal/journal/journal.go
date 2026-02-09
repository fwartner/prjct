package journal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fwartner/prjct/internal/config"
)

// OpType describes the kind of operation recorded.
type OpType string

const (
	OpCreate  OpType = "create"
	OpRename  OpType = "rename"
	OpArchive OpType = "archive"
	OpImport  OpType = "import"
	OpClone   OpType = "clone"
	OpSync    OpType = "sync"
	OpClean   OpType = "clean"
	OpNote    OpType = "note"
)

// Record represents a single journaled operation.
type Record struct {
	Timestamp time.Time         `json:"timestamp"`
	Operation OpType            `json:"operation"`
	Details   map[string]string `json:"details"`
}

// Journal holds a list of operation records.
type Journal struct {
	Records []Record `json:"records"`
}

// JournalPath returns the path to the journal file alongside the config.
func JournalPath() (string, error) {
	cfgPath, err := config.DefaultPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(cfgPath), "journal.json"), nil
}

// Load reads the journal from disk. Returns empty journal if file doesn't exist.
func Load(path string) (*Journal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Journal{}, nil
		}
		return nil, fmt.Errorf("cannot read journal: %w", err)
	}

	var j Journal
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, fmt.Errorf("corrupt journal: %w", err)
	}
	return &j, nil
}

// Save writes the journal to disk.
func Save(path string, j *Journal) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create journal directory: %w", err)
	}

	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal journal: %w", err)
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0644)
}

// Append adds a record to the journal.
func Append(path string, rec Record) error {
	j, err := Load(path)
	if err != nil {
		return err
	}
	j.Records = append(j.Records, rec)

	// Keep only last 100 records
	if len(j.Records) > 100 {
		j.Records = j.Records[len(j.Records)-100:]
	}

	return Save(path, j)
}

// Last returns the most recent record, or nil if empty.
func Last(path string) (*Record, error) {
	j, err := Load(path)
	if err != nil {
		return nil, err
	}
	if len(j.Records) == 0 {
		return nil, nil
	}
	rec := j.Records[len(j.Records)-1]
	return &rec, nil
}

// RemoveLast removes the most recent record from the journal.
func RemoveLast(path string) error {
	j, err := Load(path)
	if err != nil {
		return err
	}
	if len(j.Records) == 0 {
		return nil
	}
	j.Records = j.Records[:len(j.Records)-1]
	return Save(path, j)
}
