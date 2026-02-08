package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available project templates",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	fmt.Println("Available templates:")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "  ID\tNAME\tBASE PATH\n")
	fmt.Fprintf(w, "  --\t----\t---------\n")
	for _, t := range cfg.Templates {
		fmt.Fprintf(w, "  %s\t%s\t%s\n", t.ID, t.Name, t.BasePath)
	}
	w.Flush()

	return nil
}
