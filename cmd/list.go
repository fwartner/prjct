package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listTag string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available project templates",
	Long:  `Lists all configured templates. Use --tag to filter by tag.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVarP(&listTag, "tag", "t", "", "filter templates by tag")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	var tags []string
	if listTag != "" {
		tags = []string{listTag}
	}

	fmt.Println("Available templates:")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "  ID\tNAME\tBASE PATH\tTAGS\n")
	fmt.Fprintf(w, "  --\t----\t---------\t----\n")
	count := 0
	for _, t := range cfg.Templates {
		if !t.MatchesTags(tags) {
			continue
		}
		tagStr := strings.Join(t.Tags, ", ")
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", t.ID, t.Name, t.BasePath, tagStr)
		count++
	}
	w.Flush()

	if listTag != "" && count == 0 {
		fmt.Printf("\nNo templates matching tag %q\n", listTag)
	}

	return nil
}
