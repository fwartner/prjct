package cmd

import (
	"fmt"

	"github.com/fwartner/prjct/internal/config"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate a template YAML file",
	Long: `Parses and validates a YAML file containing templates without
importing it. Useful for checking templates before sharing or importing.`,
	Args: cobra.ExactArgs(1),
	RunE: runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(args[0])
	if err != nil {
		return &ExitError{Code: ExitConfigInvalid, Message: fmt.Sprintf("parse error: %v", err)}
	}

	errs := cfg.Validate()
	if len(errs) == 0 {
		fmt.Printf("Valid: %d template(s) found\n", len(cfg.Templates))
		for _, t := range cfg.Templates {
			fmt.Printf("  - %s (%s)\n", t.Name, t.ID)
		}
		return nil
	}

	fmt.Printf("Found %d issue(s):\n", len(errs))
	for _, e := range errs {
		fmt.Printf("  - %s\n", e.Error())
	}
	return &ExitError{Code: ExitConfigInvalid, Message: fmt.Sprintf("validation failed: %d issue(s)", len(errs))}
}
