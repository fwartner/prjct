package cmd

import (
	"fmt"

	"github.com/fwartner/prjct/internal/config"
	"github.com/spf13/cobra"
)

var treeCmd = &cobra.Command{
	Use:   "tree <template-id>",
	Short: "Preview a template's directory structure",
	Args:  cobra.ExactArgs(1),
	RunE:  runTree,
}

func runTree(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	tmpl, err := cfg.ResolveTemplate(args[0])
	if err != nil {
		return &ExitError{
			Code:    ExitTemplateNotFound,
			Message: err.Error(),
		}
	}

	fmt.Printf("%s (%s)\n", tmpl.Name, tmpl.ID)
	printTree(tmpl.Directories, "")
	return nil
}

func printTree(dirs []config.Directory, prefix string) {
	for i, d := range dirs {
		isLast := i == len(dirs)-1

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		label := d.Name
		if d.Optional {
			label += " (optional)"
		}
		fmt.Printf("%s%s%s\n", prefix, connector, label)

		// Print files
		childPrefix := prefix + "│   "
		if isLast {
			childPrefix = prefix + "    "
		}

		for j, f := range d.Files {
			fileIsLast := j == len(d.Files)-1 && len(d.Children) == 0
			fileConnector := "├── "
			if fileIsLast {
				fileConnector = "└── "
			}
			fmt.Printf("%s%s📄 %s\n", childPrefix, fileConnector, f.Name)
		}

		if len(d.Children) > 0 {
			printTree(d.Children, childPrefix)
		}
	}
}
