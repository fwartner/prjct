package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show project statistics",
	Long:  `Displays statistics about indexed projects grouped by template.`,
	RunE:  runStats,
}

func runStats(cmd *cobra.Command, args []string) error {
	idxPath, err := resolveIndexPath()
	if err != nil {
		return err
	}

	idx, err := index.Load(idxPath)
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("cannot load index: %v", err)}
	}

	if len(idx.Projects) == 0 {
		fmt.Println("No projects indexed yet.")
		return nil
	}

	// Count per template
	counts := make(map[string]int)
	names := make(map[string]string) // id → name
	for _, e := range idx.Projects {
		counts[e.TemplateID]++
		if e.TemplateName != "" {
			names[e.TemplateID] = e.TemplateName
		}
	}

	fmt.Printf("Total projects: %d\n\n", len(idx.Projects))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "  TEMPLATE\tNAME\tCOUNT\n")
	fmt.Fprintf(w, "  --------\t----\t-----\n")
	for id, count := range counts {
		name := names[id]
		if name == "" {
			name = "-"
		}
		fmt.Fprintf(w, "  %s\t%s\t%d\n", id, name, count)
	}
	w.Flush()

	return nil
}
