package cmd

import (
	"fmt"
	"os"

	"github.com/fwartner/prjct/internal/config"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate configuration",
	Long:  `Checks the config file for existence, syntax, and semantic correctness.`,
	RunE:  runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	path := configPath
	if path == "" {
		var err error
		path, err = config.DefaultPath()
		if err != nil {
			return &ExitError{Code: ExitGeneral, Message: err.Error()}
		}
	}

	fmt.Printf("Checking config: %s\n\n", path)

	passed := 0
	warnings := 0
	failures := 0

	// Check 1: File exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		printCheck("FAIL", "Config file exists")
		failures++
		fmt.Printf("\nResult: %d passed, %d warnings, %d errors\n", passed, warnings, failures)
		return &ExitError{
			Code:    ExitConfigNotFound,
			Message: fmt.Sprintf("config file not found: %s\nRun 'prjct install' to create a default config.", path),
		}
	}
	printCheck("OK", "Config file exists")
	passed++

	// Check 2: YAML parses
	cfg, err := config.Load(path)
	if err != nil {
		printCheck("FAIL", fmt.Sprintf("YAML syntax valid (%v)", err))
		failures++
		fmt.Printf("\nResult: %d passed, %d warnings, %d errors\n", passed, warnings, failures)
		return &ExitError{
			Code:    ExitConfigInvalid,
			Message: fmt.Sprintf("config parse error: %v", err),
		}
	}
	printCheck("OK", "YAML syntax valid")
	passed++

	// Check 3: Templates found
	if len(cfg.Templates) == 0 {
		printCheck("FAIL", "Templates found")
		failures++
	} else {
		printCheck("OK", fmt.Sprintf("%d templates found", len(cfg.Templates)))
		passed++
	}

	// Check 4: Validation
	errs := cfg.Validate()
	if len(errs) > 0 {
		for _, e := range errs {
			printCheck("FAIL", e.Error())
			failures++
		}
	} else {
		printCheck("OK", "All template IDs unique")
		passed++
		printCheck("OK", "No reserved ID conflicts")
		passed++
		printCheck("OK", "Directory trees valid")
		passed++
	}

	// Check 5: Base paths exist
	for _, t := range cfg.Templates {
		expanded, err := config.ExpandPath(t.BasePath)
		if err != nil {
			printCheck("WARN", fmt.Sprintf("Cannot expand base path for %q: %v", t.ID, err))
			warnings++
			continue
		}
		if _, err := os.Stat(expanded); os.IsNotExist(err) {
			printCheck("WARN", fmt.Sprintf("Base path does not exist: %s (template: %s)", expanded, t.ID))
			warnings++
		} else {
			printCheck("OK", fmt.Sprintf("Base path exists: %s (template: %s)", expanded, t.ID))
			passed++
		}
	}

	fmt.Printf("\nResult: %d passed, %d warnings, %d errors\n", passed, warnings, failures)

	if failures > 0 {
		return &ExitError{
			Code:    ExitConfigInvalid,
			Message: fmt.Sprintf("config has %d errors", failures),
		}
	}

	return nil
}

func printCheck(status, message string) {
	fmt.Printf("  [%-4s] %s\n", status, message)
}
