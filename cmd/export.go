package cmd

import (
	"fmt"
	"os"

	"github.com/fwartner/prjct/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export <template-id>",
	Short: "Export a template to a YAML file",
	Long:  `Exports a single template (with inheritance resolved) to a standalone YAML file.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "output file path (default: <id>.yaml)")
}

func runExport(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	tmpl, err := cfg.ResolveTemplate(args[0])
	if err != nil {
		return &ExitError{Code: ExitTemplateNotFound, Message: err.Error()}
	}

	// Clear extends since we resolved it
	tmpl.Extends = ""

	wrapped := &config.Config{Templates: []config.Template{*tmpl}}
	data, err := yaml.Marshal(wrapped)
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("marshaling template: %v", err)}
	}

	outputPath := exportOutput
	if outputPath == "" {
		outputPath = args[0] + ".yaml"
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("writing file: %v", err)}
	}

	fmt.Printf("Exported template %q to %s\n", args[0], outputPath)
	return nil
}
