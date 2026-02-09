package cmd

import (
	"fmt"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
	Use:   "note <query> [text]",
	Short: "Add or view notes on a project",
	Long: `Without text, shows existing notes for the first matching project.
With text, appends a note to the project's index entry.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runNote,
}

func runNote(cmd *cobra.Command, args []string) error {
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

	// View mode
	if len(args) == 1 {
		fmt.Printf("Notes for %s (%s):\n", entry.Name, entry.Path)
		if len(entry.Notes) == 0 {
			fmt.Println("  (no notes)")
			return nil
		}
		for i, n := range entry.Notes {
			fmt.Printf("  %d. %s\n", i+1, n)
		}
		return nil
	}

	// Add mode
	noteText := args[1]
	err = index.Update(idxPath, entry.Path, func(e *index.Entry) {
		e.Notes = append(e.Notes, noteText)
	})
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("failed to save note: %v", err)}
	}

	fmt.Printf("Note added to %s\n", entry.Name)
	return nil
}
