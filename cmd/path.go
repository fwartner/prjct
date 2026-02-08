package cmd

import (
	"fmt"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

var pathTemplate string

var pathCmd = &cobra.Command{
	Use:   "path <query>",
	Short: "Print the path of a matching project",
	Long:  `Searches the project index and prints the path of the first match. Useful for scripting: cd $(prjct path myproject)`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPath,
}

func init() {
	pathCmd.Flags().StringVarP(&pathTemplate, "template", "t", "", "filter by template ID")
}

func runPath(cmd *cobra.Command, args []string) error {
	idxPath, err := resolveIndexPath()
	if err != nil {
		return err
	}

	idx, err := index.Load(idxPath)
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("cannot load index: %v", err)}
	}

	results := index.Search(idx, args[0])
	if pathTemplate != "" {
		results = index.FilterByTemplate(results, pathTemplate)
	}

	if len(results) == 0 {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("no project matching %q", args[0])}
	}

	fmt.Println(results[0].Path)
	return nil
}
