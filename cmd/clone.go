package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fwartner/prjct/internal/index"
	"github.com/fwartner/prjct/internal/project"
	"github.com/spf13/cobra"
)

var cloneWithFiles bool

var cloneCmd = &cobra.Command{
	Use:   "clone <query> <new-name>",
	Short: "Clone a project's directory structure",
	Long: `Creates a new project by duplicating the directory structure of an
existing project. Use --with-files to also copy file contents.`,
	Args: cobra.ExactArgs(2),
	RunE: runClone,
}

func init() {
	cloneCmd.Flags().BoolVar(&cloneWithFiles, "with-files", false, "also copy files")
}

func runClone(cmd *cobra.Command, args []string) error {
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
	sourcePath := entry.Path

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("source directory not found: %s", sourcePath)}
	}

	newName, err := project.Sanitize(args[1])
	if err != nil {
		return &ExitError{Code: ExitInvalidName, Message: fmt.Sprintf("invalid name: %v", err)}
	}

	destPath := filepath.Join(filepath.Dir(sourcePath), newName)
	if _, err := os.Stat(destPath); err == nil {
		return &ExitError{Code: ExitProjectExists, Message: fmt.Sprintf("destination already exists: %s", destPath)}
	}

	if dryRun {
		fmt.Printf("Dry run — would clone %s to %s\n", sourcePath, destPath)
		return nil
	}

	dirCount := 0
	fileCount := 0

	err = filepath.WalkDir(sourcePath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, relErr := filepath.Rel(sourcePath, path)
		if relErr != nil {
			return relErr
		}

		target := filepath.Join(destPath, rel)

		if d.IsDir() {
			if mkErr := os.MkdirAll(target, 0755); mkErr != nil {
				return mkErr
			}
			dirCount++
			if verbose {
				fmt.Printf("  mkdir %s\n", target)
			}
			return nil
		}

		if cloneWithFiles {
			if cpErr := copyFile(path, target); cpErr != nil {
				return cpErr
			}
			fileCount++
			if verbose {
				fmt.Printf("  copy  %s\n", target)
			}
		}
		return nil
	})

	if err != nil {
		return &ExitError{Code: ExitCreateFailed, Message: fmt.Sprintf("clone failed: %v", err)}
	}

	// Best-effort index update
	_ = index.Add(idxPath, index.Entry{
		Name:         newName,
		TemplateID:   entry.TemplateID,
		TemplateName: entry.TemplateName,
		Path:         destPath,
		CreatedAt:    time.Now(),
	})

	fmt.Printf("Cloned: %s\n", sourcePath)
	fmt.Printf("    To: %s\n", destPath)
	fmt.Printf("  Dirs: %d\n", dirCount)
	if fileCount > 0 {
		fmt.Printf(" Files: %d\n", fileCount)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
