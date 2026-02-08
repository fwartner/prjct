package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fwartner/prjct/internal/config"
	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

var reindexTemplate string

var reindexCmd = &cobra.Command{
	Use:   "reindex",
	Short: "Discover and index existing projects from template base paths",
	Long: `Scans template base directories for existing project folders and adds
them to the search index. Use this to index projects created before
the search feature was available or created outside of prjct.`,
	RunE: runReindex,
}

func init() {
	reindexCmd.Flags().StringVarP(&reindexTemplate, "template", "t", "", "only scan a specific template's base path")
}

func runReindex(cmd *cobra.Command, args []string) error {
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

	// Build set of existing paths for fast duplicate check
	existing := make(map[string]bool, len(idx.Projects))
	for _, e := range idx.Projects {
		existing[e.Path] = true
	}

	templates := cfg.Templates
	if reindexTemplate != "" {
		t := cfg.FindTemplate(reindexTemplate)
		if t == nil {
			return &ExitError{
				Code:    ExitTemplateNotFound,
				Message: fmt.Sprintf("template %q not found", reindexTemplate),
			}
		}
		templates = []config.Template{*t}
	}

	added := 0
	scanned := 0
	for _, tmpl := range templates {
		expanded, err := config.ExpandPath(tmpl.BasePath)
		if err != nil {
			if verbose {
				fmt.Printf("  skip %s: %v\n", tmpl.ID, err)
			}
			continue
		}

		entries, err := os.ReadDir(expanded)
		if err != nil {
			if verbose {
				fmt.Printf("  skip %s: %v\n", tmpl.ID, err)
			}
			continue
		}

		scanned++
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			projectPath := filepath.Join(expanded, entry.Name())
			if existing[projectPath] {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			idx.Projects = append(idx.Projects, index.Entry{
				Name:         entry.Name(),
				TemplateID:   tmpl.ID,
				TemplateName: tmpl.Name,
				Path:         projectPath,
				CreatedAt:    info.ModTime(),
			})
			existing[projectPath] = true
			added++

			if verbose {
				fmt.Printf("  + %s (%s)\n", entry.Name(), tmpl.ID)
			}
		}
	}

	if err := index.Save(idxPath, idx); err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("cannot save index: %v", err)}
	}

	total := len(idx.Projects)
	fmt.Printf("Indexed %d new project(s) across %d template(s) (%d total)\n", added, scanned, total)

	return nil
}

// reindexTime is used for testing; defaults to time.Now
var reindexTime = func() time.Time { return time.Now() }
