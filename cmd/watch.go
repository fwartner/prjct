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

var watchInterval int

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch base paths and auto-index new projects",
	Long: `Periodically scans all template base paths for new directories and
adds them to the project index. Press Ctrl+C to stop.`,
	Args: cobra.NoArgs,
	RunE: runWatch,
}

func init() {
	watchCmd.Flags().IntVar(&watchInterval, "interval", 30, "scan interval in seconds")
}

func runWatch(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	idxPath, err := resolveIndexPath()
	if err != nil {
		return err
	}

	if watchInterval < 1 {
		watchInterval = 30
	}

	fmt.Printf("Watching base paths every %ds (Ctrl+C to stop)...\n", watchInterval)

	// Run once immediately, then on interval
	scanAndIndex(cfg, idxPath)

	ticker := time.NewTicker(time.Duration(watchInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		scanAndIndex(cfg, idxPath)
	}

	return nil
}

func scanAndIndex(cfg *config.Config, idxPath string) {
	idx, err := index.Load(idxPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot load index: %v\n", err)
		return
	}

	existing := make(map[string]bool, len(idx.Projects))
	for _, e := range idx.Projects {
		existing[e.Path] = true
	}

	added := 0
	for _, tmpl := range cfg.Templates {
		basePath, err := config.ExpandPath(tmpl.BasePath)
		if err != nil {
			continue
		}

		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			fullPath := filepath.Join(basePath, entry.Name())
			if existing[fullPath] {
				continue
			}

			info, infoErr := entry.Info()
			createdAt := time.Now()
			if infoErr == nil {
				createdAt = info.ModTime()
			}

			if err := index.Add(idxPath, index.Entry{
				Name:         entry.Name(),
				TemplateID:   tmpl.ID,
				TemplateName: tmpl.Name,
				Path:         fullPath,
				CreatedAt:    createdAt,
			}); err == nil {
				existing[fullPath] = true
				added++
				if verbose {
					fmt.Printf("  indexed: %s (%s)\n", entry.Name(), tmpl.ID)
				}
			}
		}
	}

	if added > 0 {
		fmt.Printf("[%s] Indexed %d new project(s)\n", time.Now().Format("15:04:05"), added)
	}
}
