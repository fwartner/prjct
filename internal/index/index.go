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
	Notes        []string  `json:"notes,omitempty"`
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

// FuzzySearch returns entries where any searchable field is within the given
// edit-distance threshold of the query. Falls back to substring matching first.
func FuzzySearch(idx *Index, query string, maxDist int) []Entry {
	if query == "" {
		return idx.Projects
	}

	// First try exact substring match
	results := Search(idx, query)
	if len(results) > 0 {
		return results
	}

	// Fuzzy match using Levenshtein distance on individual fields
	q := strings.ToLower(query)
	for _, e := range idx.Projects {
		fields := []string{
			strings.ToLower(e.Name),
			strings.ToLower(e.TemplateID),
			strings.ToLower(e.TemplateName),
		}
		for _, f := range fields {
			if levenshtein(q, f) <= maxDist || containsFuzzy(f, q, maxDist) {
				results = append(results, e)
				break
			}
		}
	}
	return results
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = minInt(curr[j-1]+1, minInt(prev[j]+1, prev[j-1]+cost))
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

// containsFuzzy checks if any substring of haystack of length len(needle)±maxDist
// is within maxDist edits of needle.
func containsFuzzy(haystack, needle string, maxDist int) bool {
	if len(needle) > len(haystack)+maxDist {
		return false
	}
	// Slide a window and check distance
	windowSize := len(needle)
	for start := 0; start <= len(haystack)-windowSize+maxDist && start < len(haystack); start++ {
		end := start + windowSize + maxDist
		if end > len(haystack) {
			end = len(haystack)
		}
		sub := haystack[start:end]
		if levenshtein(needle, sub) <= maxDist {
			return true
		}
	}
	return false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
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
