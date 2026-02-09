package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

var (
	searchTemplate string
	searchFuzzy    bool
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search indexed projects",
	Long: `Search for previously created projects by name, template, or path.
Run without a query to list all indexed projects.
Use --template to filter by template ID.
Use --fuzzy for approximate matching.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().StringVarP(&searchTemplate, "template", "t", "", "filter by template ID")
	searchCmd.Flags().BoolVar(&searchFuzzy, "fuzzy", false, "enable fuzzy (approximate) matching")
}

func runSearch(cmd *cobra.Command, args []string) error {
	idxPath, err := resolveIndexPath()
	if err != nil {
		return err
	}

	idx, err := index.Load(idxPath)
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("cannot load project index: %v", err)}
	}

	query := ""
	if len(args) == 1 {
		query = args[0]
	}

	var results []index.Entry
	if searchFuzzy && query != "" {
		results = index.FuzzySearch(idx, query, 2)
	} else {
		results = index.Search(idx, query)
	}

	if searchTemplate != "" {
		results = index.FilterByTemplate(results, searchTemplate)
	}

	if len(results) == 0 {
		fmt.Println("Found 0 project(s)")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "  NAME\tTEMPLATE\tPATH\tCREATED\n")
	fmt.Fprintf(w, "  ----\t--------\t----\t-------\n")
	for _, e := range results {
		created := e.CreatedAt.Format("2006-01-02 15:04")
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", e.Name, e.TemplateID, e.Path, created)
	}
	w.Flush()

	fmt.Printf("\nFound %d project(s)\n", len(results))
	return nil
}

func resolveIndexPath() (string, error) {
	// If --config is set, derive index path from its directory
	if configPath != "" {
		return filepath.Join(filepath.Dir(configPath), "projects.json"), nil
	}
	p, err := index.IndexPath()
	if err != nil {
		return "", &ExitError{Code: ExitGeneral, Message: err.Error()}
	}
	return p, nil
}
