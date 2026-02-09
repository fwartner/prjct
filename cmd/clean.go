package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean <query>",
	Short: "Remove empty directories from a project",
	Long: `Searches the project index and removes all empty directories
from the first matching project. Operates recursively — a directory
that becomes empty after its children are removed is also deleted.`,
	Args: cobra.ExactArgs(1),
	RunE: runClean,
}

func runClean(cmd *cobra.Command, args []string) error {
	idxPath, err := resolveIndexPath()
	if err != nil {
		return err
	}

	idx, err := index.Load(idxPath)
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("cannot load index: %v", err)}
	}

	results := index.Search(idx, args[0])
	if len(results) == 0 {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("no project matching %q", args[0])}
	}

	entry := results[0]
	projectPath := entry.Path

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("project directory not found: %s", projectPath)}
	}

	// Collect all directories, deepest first
	var dirs []string
	_ = filepath.WalkDir(projectPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || !d.IsDir() {
			return nil
		}
		if path == projectPath {
			return nil // don't remove the root
		}
		dirs = append(dirs, path)
		return nil
	})

	// Sort deepest first so children are removed before parents
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})

	removed := 0
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		if len(entries) > 0 {
			continue
		}

		if dryRun {
			fmt.Printf("  [DRY-RUN] rmdir %s\n", dir)
			removed++
			continue
		}
		if err := os.Remove(dir); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "  Warning: cannot remove %s: %v\n", dir, err)
			}
			continue
		}
		if verbose {
			fmt.Printf("  rmdir %s\n", dir)
		}
		removed++
	}

	if dryRun {
		fmt.Printf("Dry run — would remove %d empty directory(ies)\n", removed)
	} else {
		fmt.Printf("Removed %d empty directory(ies)\n", removed)
	}
	return nil
}
