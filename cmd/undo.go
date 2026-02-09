package cmd

import (
	"fmt"
	"os"

	"github.com/fwartner/prjct/internal/index"
	"github.com/fwartner/prjct/internal/journal"
	"github.com/spf13/cobra"
)

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo the last recorded operation",
	Long: `Reverses the most recent journaled operation such as create, rename,
or clone. Not all operations can be fully undone.`,
	Args: cobra.NoArgs,
	RunE: runUndo,
}

func runUndo(cmd *cobra.Command, args []string) error {
	jPath, err := resolveJournalPath()
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: err.Error()}
	}

	rec, err := journal.Last(jPath)
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("cannot read journal: %v", err)}
	}
	if rec == nil {
		fmt.Println("Nothing to undo.")
		return nil
	}

	fmt.Printf("Last operation: %s at %s\n", rec.Operation, rec.Timestamp.Format("2006-01-02 15:04:05"))
	for k, v := range rec.Details {
		fmt.Printf("  %s: %s\n", k, v)
	}

	if dryRun {
		fmt.Println("Dry run — would attempt to undo the above operation.")
		return nil
	}

	switch rec.Operation {
	case journal.OpCreate, journal.OpClone:
		path := rec.Details["path"]
		if path == "" {
			return &ExitError{Code: ExitGeneral, Message: "journal record missing path"}
		}
		if err := os.RemoveAll(path); err != nil {
			return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("cannot remove %s: %v", path, err)}
		}
		// Remove from index
		if idxPath, idxErr := resolveIndexPath(); idxErr == nil {
			_ = index.Remove(idxPath, path)
		}
		fmt.Printf("Removed: %s\n", path)

	case journal.OpRename:
		oldPath := rec.Details["old_path"]
		newPath := rec.Details["new_path"]
		if oldPath == "" || newPath == "" {
			return &ExitError{Code: ExitGeneral, Message: "journal record missing paths"}
		}
		if err := os.Rename(newPath, oldPath); err != nil {
			return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("cannot rename back: %v", err)}
		}
		// Update index
		if idxPath, idxErr := resolveIndexPath(); idxErr == nil {
			oldName := rec.Details["old_name"]
			_ = index.Update(idxPath, newPath, func(e *index.Entry) {
				e.Name = oldName
				e.Path = oldPath
			})
		}
		fmt.Printf("Reverted: %s -> %s\n", newPath, oldPath)

	default:
		fmt.Printf("Cannot undo operation type %q automatically.\n", rec.Operation)
		return nil
	}

	// Remove the record from journal
	_ = journal.RemoveLast(jPath)
	fmt.Println("Undo complete.")
	return nil
}

func resolveJournalPath() (string, error) {
	if configPath != "" {
		return configPath[:len(configPath)-len("config.yaml")] + "journal.json", nil
	}
	return journal.JournalPath()
}
