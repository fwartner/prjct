package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fwartner/prjct/internal/index"
	"github.com/fwartner/prjct/internal/project"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync <query>",
	Short: "Sync a project with its template",
	Long: `Compares a project against its template and creates any missing
directories. Uses the project index to find the template that was used.`,
	Args: cobra.ExactArgs(1),
	RunE: runSync,
}

func runSync(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	idxPath, err := resolveIndexPath()
	if err != nil {
		return err
	}

	idx, err := index.Load(idxPath)
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("cannot load index: %v", err)}
	}

	results := index.Search(idx, args[0])
	if len(results) == 0 {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("no project matching %q", args[0])}
	}

	entry := results[0]
	projectPath := entry.Path

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("project directory not found: %s", projectPath)}
	}

	tmpl, err := cfg.ResolveTemplate(entry.TemplateID)
	if err != nil {
		return &ExitError{Code: ExitTemplateNotFound, Message: fmt.Sprintf("template %q not found in config", entry.TemplateID)}
	}

	// Get template dirs
	templateDirs := make(map[string]bool)
	for _, p := range project.Flatten(tmpl.Directories, "") {
		templateDirs[filepath.ToSlash(p)] = true
	}

	// Get actual dirs
	actualDirs := make(map[string]bool)
	_ = filepath.WalkDir(projectPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(projectPath, path)
		if relErr != nil || rel == "." {
			return nil
		}
		actualDirs[filepath.ToSlash(rel)] = true
		return nil
	})

	// Find missing dirs
	var missing []string
	for p := range templateDirs {
		if !actualDirs[p] {
			missing = append(missing, p)
		}
	}
	sort.Strings(missing)

	if len(missing) == 0 {
		fmt.Println("Project is in sync with template — no missing directories.")
		return nil
	}

	created := 0
	for _, rel := range missing {
		fullPath := filepath.Join(projectPath, filepath.FromSlash(rel))
		if dryRun {
			fmt.Printf("  [DRY-RUN] mkdir %s\n", fullPath)
			created++
			continue
		}
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: cannot create %s: %v\n", rel, err)
			continue
		}
		if verbose {
			fmt.Printf("  mkdir %s\n", fullPath)
		}
		created++
	}

	if dryRun {
		fmt.Printf("\nDry run — would create %d missing directory(ies)\n", created)
	} else {
		fmt.Printf("Synced %d missing directory(ies)\n", created)
	}
	return nil
}
