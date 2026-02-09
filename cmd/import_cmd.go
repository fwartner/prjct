package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fwartner/prjct/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var importCmd = &cobra.Command{
	Use:   "import <file-or-url>",
	Short: "Import templates from a YAML file or URL",
	Long: `Imports templates from a local YAML file or a remote URL into the
current configuration. Skips templates with conflicting IDs.`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func runImport(cmd *cobra.Command, args []string) error {
	source := args[0]

	var data []byte
	var err error

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		data, err = fetchURL(source)
		if err != nil {
			return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("fetching URL: %v", err)}
		}
	} else {
		data, err = os.ReadFile(source)
		if err != nil {
			return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("reading file: %v", err)}
		}
	}

	var imported config.Config
	if err := yaml.Unmarshal(data, &imported); err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("parsing import: %v", err)}
	}

	if len(imported.Templates) == 0 {
		return &ExitError{Code: ExitGeneral, Message: "import source contains no templates"}
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

func fetchURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}
