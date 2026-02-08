package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

var recentCmd = &cobra.Command{
	Use:   "recent [n]",
	Short: "Show recently created projects",
	Long:  `Lists the most recently created projects. Defaults to 10.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRecent,
}

func runRecent(cmd *cobra.Command, args []string) error {
	n := 10
	if len(args) == 1 {
		var err error
		n, err = strconv.Atoi(args[0])
		if err != nil || n < 1 {
			return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("invalid count %q", args[0])}
		}
	}

	idxPath, err := resolveIndexPath()
	if err != nil {
		return err
	}

	idx, err := index.Load(idxPath)
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("cannot load index: %v", err)}
	}

	entries := idx.Projects
	index.SortByCreatedDesc(entries)
	if n < len(entries) {
		entries = entries[:n]
	}

	if len(entries) == 0 {
		fmt.Println("No projects indexed yet.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "  NAME\tTEMPLATE\tPATH\tCREATED\n")
	fmt.Fprintf(w, "  ----\t--------\t----\t-------\n")
	for _, e := range entries {
		created := e.CreatedAt.Format("2006-01-02 15:04")
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", e.Name, e.TemplateID, e.Path, created)
	}
	w.Flush()

	fmt.Printf("\nShowing %d of %d project(s)\n", len(entries), len(idx.Projects))
	return nil
}
