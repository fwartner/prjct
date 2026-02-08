package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fwartner/prjct/internal/index"
	"github.com/fwartner/prjct/internal/project"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <query> <new-name>",
	Short: "Rename an existing project",
	Long: `Searches the project index for the first match and renames the
project directory on disk and in the index.`,
	Args: cobra.ExactArgs(2),
	RunE: runRename,
}

func runRename(cmd *cobra.Command, args []string) error {
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
	newName, err := project.Sanitize(args[1])
	if err != nil {
		return &ExitError{Code: ExitInvalidName, Message: fmt.Sprintf("invalid name: %v", err)}
	}

	oldPath := entry.Path
	newPath := filepath.Join(filepath.Dir(oldPath), newName)

	if _, err := os.Stat(newPath); err == nil {
		return &ExitError{Code: ExitProjectExists, Message: fmt.Sprintf("directory already exists: %s", newPath)}
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("rename failed: %v", err)}
	}

	if err := index.Update(idxPath, oldPath, func(e *index.Entry) {
		e.Name = newName
		e.Path = newPath
	}); err != nil {
		// Best effort — print warning but don't fail
		fmt.Fprintf(os.Stderr, "Warning: index update failed: %v\n", err)
	}

	fmt.Printf("Renamed: %s\n", oldPath)
	fmt.Printf("     To: %s\n", newPath)
	return nil
}
