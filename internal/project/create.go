package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fwartner/prjct/internal/config"
	tmplpkg "github.com/fwartner/prjct/internal/template"
)

// Sentinel errors for exit code mapping.
var (
	ErrProjectExists = errors.New("project directory already exists")
	ErrPermission    = errors.New("permission denied")
)

// CreateOptions holds options for project creation.
type CreateOptions struct {
	Verbose      bool
	DryRun       bool
	Variables    map[string]string
	SkipOptional map[string]bool
}

// Result holds the outcome of a project creation.
type Result struct {
	ProjectPath  string
	DirsCreated  int
	FilesCreated int
	TemplateName string
}

// ExecHook is the function used to run post-creation hooks. Replaceable for testing.
var ExecHook = execHookDefault

// Create builds the full directory tree for a project.
// On failure, it attempts to roll back all created directories and files.
func Create(tmpl *config.Template, projectName string, opts CreateOptions) (*Result, error) {
	basePath, err := config.ExpandPath(tmpl.BasePath)
	if err != nil {
		return nil, fmt.Errorf("expanding base path: %w", err)
	}

	projectRoot := filepath.Join(basePath, projectName)

	// Check if the project directory already exists
	if _, err := os.Stat(projectRoot); err == nil {
		return nil, fmt.Errorf("%w: %s", ErrProjectExists, projectRoot)
	}

	if opts.Variables == nil {
		opts.Variables = make(map[string]string)
	}
	opts.Variables["path"] = projectRoot

	// Create project root (also creates base path if needed)
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Creating: %s\n", projectRoot)
	}
	if !opts.DryRun {
		if err := os.MkdirAll(projectRoot, 0755); err != nil {
			if os.IsPermission(err) {
				return nil, fmt.Errorf("%w: %s", ErrPermission, err)
			}
			return nil, fmt.Errorf("creating project root: %w", err)
		}
	}

	created := []string{projectRoot}
	createdFiles := []string{}
	var dirCount, fileCount int
	var createErr error
	dirCount, fileCount, createErr = createTree(tmpl.Directories, projectRoot, opts, &created, &createdFiles)
	dirCount++ // add root
	if createErr != nil {
		if !opts.DryRun {
			rollback(created, createdFiles, opts.Verbose)
		}
		return nil, createErr
	}

	// Execute hooks
	if !opts.DryRun && len(tmpl.Hooks) > 0 {
		for _, hook := range tmpl.Hooks {
			resolved := tmplpkg.Resolve(hook, opts.Variables)
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "  hook: %s\n", resolved)
			}
			if err := ExecHook(resolved, projectRoot); err != nil {
				return nil, fmt.Errorf("hook %q failed: %w", hook, err)
			}
		}
	}

	return &Result{
		ProjectPath:  projectRoot,
		DirsCreated:  dirCount,
		FilesCreated: fileCount,
		TemplateName: tmpl.Name,
	}, nil
}

// createTree recursively creates directories and files, returning counts.
func createTree(dirs []config.Directory, parentPath string, opts CreateOptions, created *[]string, createdFiles *[]string) (int, int, error) {
	dirCount := 0
	fileCount := 0

	for _, d := range dirs {
		dirName := tmplpkg.Resolve(d.Name, opts.Variables)

		if d.Optional && opts.SkipOptional != nil && opts.SkipOptional[d.Name] {
			continue
		}

		// Evaluate conditional directory
		if d.When != "" && !config.EvalWhen(d.When, opts.Variables) {
			continue
		}

		fullPath := filepath.Join(parentPath, dirName)
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "  mkdir %s\n", fullPath)
		}

		if !opts.DryRun {
			if err := os.Mkdir(fullPath, 0755); err != nil {
				if os.IsPermission(err) {
					return dirCount, fileCount, fmt.Errorf("%w: %s", ErrPermission, err)
				}
				return dirCount, fileCount, fmt.Errorf("creating directory %s: %w", dirName, err)
			}
			*created = append(*created, fullPath)
		}
		dirCount++

		// Create files
		for _, f := range d.Files {
			fileName := tmplpkg.Resolve(f.Name, opts.Variables)
			filePath := filepath.Join(fullPath, fileName)
			content := tmplpkg.Resolve(f.Content, opts.Variables)

			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "  touch %s\n", filePath)
			}

			if !opts.DryRun {
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					return dirCount, fileCount, fmt.Errorf("creating file %s: %w", fileName, err)
				}
				*createdFiles = append(*createdFiles, filePath)
			}
			fileCount++
		}

		// Recurse into children
		if len(d.Children) > 0 {
			cd, cf, err := createTree(d.Children, fullPath, opts, created, createdFiles)
			if err != nil {
				return dirCount + cd, fileCount + cf, err
			}
			dirCount += cd
			fileCount += cf
		}
	}

	return dirCount, fileCount, nil
}

// Flatten recursively converts a Directory tree into a flat list of
// relative paths, parent-before-child ordering. Exported for use by diff.
func Flatten(dirs []config.Directory, prefix string) []string {
	var paths []string
	for _, d := range dirs {
		rel := d.Name
		if prefix != "" {
			rel = filepath.Join(prefix, d.Name)
		}
		paths = append(paths, rel)
		if len(d.Children) > 0 {
			paths = append(paths, Flatten(d.Children, rel)...)
		}
	}
	return paths
}

// rollback removes files then directories in reverse creation order (best-effort).
func rollback(created []string, createdFiles []string, verbose bool) {
	if verbose {
		fmt.Fprintln(os.Stderr, "Rolling back created files and directories...")
	}
	for i := len(createdFiles) - 1; i >= 0; i-- {
		err := os.Remove(createdFiles[i])
		if verbose && err != nil {
			fmt.Fprintf(os.Stderr, "  rollback: failed to remove %s: %v\n", createdFiles[i], err)
		}
	}
	for i := len(created) - 1; i >= 0; i-- {
		err := os.Remove(created[i])
		if verbose && err != nil {
			fmt.Fprintf(os.Stderr, "  rollback: failed to remove %s: %v\n", created[i], err)
		}
	}
}
