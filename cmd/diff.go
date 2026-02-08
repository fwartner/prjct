package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fwartner/prjct/internal/project"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <template-id> <project-path>",
	Short: "Compare a project against its template",
	Long: `Shows the differences between a template's expected directory structure
and the actual directories in an existing project.`,
	Args: cobra.ExactArgs(2),
	RunE: runDiff,
}

func runDiff(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	tmpl, err := cfg.ResolveTemplate(args[0])
	if err != nil {
		return &ExitError{Code: ExitTemplateNotFound, Message: err.Error()}
	}

	projectPath := args[1]
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("project path not found: %s", projectPath)}
	}

	// Get template dirs
	templateDirs := make(map[string]bool)
	for _, p := range project.Flatten(tmpl.Directories, "") {
		templateDirs[filepath.ToSlash(p)] = true
	}

	// Get actual dirs
	actualDirs := make(map[string]bool)
	err = filepath.WalkDir(projectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if !d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(projectPath, path)
		if relErr != nil || rel == "." {
			return nil
		}
		actualDirs[filepath.ToSlash(rel)] = true
		return nil
	})
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("walking project: %v", err)}
	}

	// Compare
	var missing, extra, matching []string

	for p := range templateDirs {
		if actualDirs[p] {
			matching = append(matching, p)
		} else {
			missing = append(missing, p)
		}
	}

	for p := range actualDirs {
		if !templateDirs[p] {
			extra = append(extra, p)
		}
	}

	sort.Strings(missing)
	sort.Strings(extra)
	sort.Strings(matching)

	for _, p := range missing {
		fmt.Printf("  [MISSING] %s\n", p)
	}
	for _, p := range extra {
		fmt.Printf("  [EXTRA]   %s\n", p)
	}

	fmt.Printf("\nTemplate: %d dirs | Project: %d dirs | Matching: %d | Missing: %d | Extra: %d\n",
		len(templateDirs), len(actualDirs), len(matching), len(missing), len(extra))

	return nil
}
