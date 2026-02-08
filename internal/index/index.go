package index

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fwartner/prjct/internal/config"
)

// Entry represents a single indexed project.
type Entry struct {
	Name         string    `json:"name"`
	TemplateID   string    `json:"template_id"`
	TemplateName string    `json:"template_name"`
	Path         string    `json:"path"`
	CreatedAt    time.Time `json:"created_at"`
	Status       string    `json:"status,omitempty"`
}

// Index holds all tracked projects.
type Index struct {
	Projects []Entry `json:"projects"`
}

// IndexPath returns the path to the project index file,
// stored alongside the config file.
func IndexPath() (string, error) {
	cfgPath, err := config.DefaultPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(cfgPath), "projects.json"), nil
}

// Load reads the index from disk. Returns an empty index if the file
// does not exist. Returns an error only for corrupt or unreadable files.
func Load(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Index{}, nil
		}
		return nil, fmt.Errorf("cannot read index: %w", err)
	}

	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("corrupt index file: %w", err)
	}
	return &idx, nil
}

// Save writes the index to disk with readable formatting.
func Save(path string, idx *Index) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create index directory: %w", err)
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal index: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("cannot write index: %w", err)
	}
	return nil
}

// Add appends an entry to the index. If an entry with the same Path
// already exists, it is skipped (no duplicates).
func Add(path string, entry Entry) error {
	idx, err := Load(path)
	if err != nil {
		return err
	}

	for _, e := range idx.Projects {
		if e.Path == entry.Path {
			return nil // already tracked
		}
	}

	idx.Projects = append(idx.Projects, entry)
	return Save(path, idx)
}

// Remove deletes the entry matching projectPath from the index.
// It is a no-op if the path is not found.
func Remove(path string, projectPath string) error {
	idx, err := Load(path)
	if err != nil {
		return err
	}

	filtered := idx.Projects[:0]
	for _, e := range idx.Projects {
		if e.Path != projectPath {
			filtered = append(filtered, e)
		}
	}
	idx.Projects = filtered
	return Save(path, idx)
}

// Search returns entries matching query as a case-insensitive substring
// of Name, TemplateID, TemplateName, or Path. An empty query returns all entries.
func Search(idx *Index, query string) []Entry {
	if query == "" {
		return idx.Projects
	}

	q := strings.ToLower(query)
	var results []Entry
	for _, e := range idx.Projects {
		if strings.Contains(strings.ToLower(e.Name), q) ||
			strings.Contains(strings.ToLower(e.TemplateID), q) ||
			strings.Contains(strings.ToLower(e.TemplateName), q) ||
			strings.Contains(strings.ToLower(e.Path), q) {
			results = append(results, e)
		}
	}
	return results
}

// FilterByTemplate returns only entries matching the given template ID.
func FilterByTemplate(entries []Entry, templateID string) []Entry {
	var results []Entry
	for _, e := range entries {
		if e.TemplateID == templateID {
			results = append(results, e)
		}
	}
	return results
}

// SortByCreatedDesc sorts entries by CreatedAt descending (newest first).
func SortByCreatedDesc(entries []Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CreatedAt.After(entries[j].CreatedAt)
	})
}

// Update modifies the entry with the given projectPath using the provided
// function. If no entry matches, it is a no-op. The index is saved to disk.
func Update(path string, projectPath string, fn func(*Entry)) error {
	idx, err := Load(path)
	if err != nil {
		return err
	}

	for i := range idx.Projects {
		if idx.Projects[i].Path == projectPath {
			fn(&idx.Projects[i])
			return Save(path, idx)
		}
	}
	return nil
}
