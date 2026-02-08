package cmd

import (
	"fmt"

	"github.com/fwartner/prjct/internal/config"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import templates from a YAML file",
	Long:  `Imports templates from a YAML file into the current configuration. Skips templates with conflicting IDs.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runImport,
}

func runImport(cmd *cobra.Command, args []string) error {
	// Load imported file
	imported, err := config.Load(args[0])
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("reading import file: %v", err)}
	}

	if len(imported.Templates) == 0 {
		return &ExitError{Code: ExitGeneral, Message: "import file contains no templates"}
	}

	// Get current config path
	cfgPath := configPath
	if cfgPath == "" {
		var pathErr error
		cfgPath, pathErr = config.DefaultPath()
		if pathErr != nil {
			return &ExitError{Code: ExitGeneral, Message: pathErr.Error()}
		}
	}

	// Load existing config
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return &ExitError{Code: ExitConfigInvalid, Message: fmt.Sprintf("loading config: %v", err)}
	}

	// Build existing ID set
	existing := make(map[string]bool)
	for _, t := range cfg.Templates {
		existing[t.ID] = true
	}

	added := 0
	skipped := 0
	for _, t := range imported.Templates {
		if existing[t.ID] {
			fmt.Printf("  skip: %q (ID conflict)\n", t.ID)
			skipped++
			continue
		}
		cfg.Templates = append(cfg.Templates, t)
		existing[t.ID] = true
		added++
		fmt.Printf("  add:  %q (%s)\n", t.ID, t.Name)
	}

	if added > 0 {
		if err := cfg.Save(cfgPath); err != nil {
			return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("saving config: %v", err)}
		}
	}

	fmt.Printf("\nImported %d template(s), skipped %d\n", added, skipped)
	return nil
}
