package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fwartner/prjct/internal/config"
)

// Sentinel errors for exit code mapping.
var (
	ErrProjectExists = errors.New("project directory already exists")
	ErrPermission    = errors.New("permission denied")
)

// Result holds the outcome of a project creation.
type Result struct {
	ProjectPath  string
	DirsCreated  int
	TemplateName string
}

// Create builds the full directory tree for a project.
// On failure, it attempts to roll back all created directories.
func Create(tmpl *config.Template, projectName string, verbose bool) (*Result, error) {
	basePath, err := config.ExpandPath(tmpl.BasePath)
	if err != nil {
		return nil, fmt.Errorf("expanding base path: %w", err)
	}

	projectRoot := filepath.Join(basePath, projectName)

	// Check if the project directory already exists
	if _, err := os.Stat(projectRoot); err == nil {
		return nil, fmt.Errorf("%w: %s", ErrProjectExists, projectRoot)
	}

	// Flatten directory tree into ordered relative paths
	paths := flatten(tmpl.Directories, "")

	// Track created directories for rollback
	created := make([]string, 0, len(paths)+1)

	// Create project root (also creates base path if needed)
	if verbose {
		fmt.Fprintf(os.Stderr, "Creating: %s\n", projectRoot)
	}
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf("%w: %s", ErrPermission, err)
		}
		return nil, fmt.Errorf("creating project root: %w", err)
	}
	created = append(created, projectRoot)

	// Create each subdirectory
	for _, relPath := range paths {
		fullPath := filepath.Join(projectRoot, relPath)
		if verbose {
			fmt.Fprintf(os.Stderr, "  mkdir %s\n", relPath)
		}
		if err := os.Mkdir(fullPath, 0755); err != nil {
			if os.IsPermission(err) {
				rollback(created, verbose)
				return nil, fmt.Errorf("%w: %s", ErrPermission, err)
			}
			rollback(created, verbose)
			return nil, fmt.Errorf("creating directory %s: %w", relPath, err)
		}
		created = append(created, fullPath)
	}

	return &Result{
		ProjectPath:  projectRoot,
		DirsCreated:  len(created),
		TemplateName: tmpl.Name,
	}, nil
}

// flatten recursively converts a Directory tree into a flat list of
// relative paths, parent-before-child ordering.
func flatten(dirs []config.Directory, prefix string) []string {
	var paths []string
	for _, d := range dirs {
		rel := d.Name
		if prefix != "" {
			rel = filepath.Join(prefix, d.Name)
		}
		paths = append(paths, rel)
		if len(d.Children) > 0 {
			paths = append(paths, flatten(d.Children, rel)...)
		}
	}
	return paths
}

// rollback removes directories in reverse creation order (best-effort).
func rollback(created []string, verbose bool) {
	if verbose {
		fmt.Fprintln(os.Stderr, "Rolling back created directories...")
	}
	for i := len(created) - 1; i >= 0; i-- {
		err := os.Remove(created[i])
		if verbose && err != nil {
			fmt.Fprintf(os.Stderr, "  rollback: failed to remove %s: %v\n", created[i], err)
		}
	}
}
