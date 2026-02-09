package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <query>",
	Short: "Show detailed information about a project",
	Long: `Displays detailed information about a project including disk size,
file and directory counts, age, template, and notes.`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("Name:     %s\n", entry.Name)
	fmt.Printf("Template: %s (%s)\n", entry.TemplateName, entry.TemplateID)
	fmt.Printf("Path:     %s\n", entry.Path)
	fmt.Printf("Created:  %s\n", entry.CreatedAt.Format("2006-01-02 15:04:05"))

	age := time.Since(entry.CreatedAt)
	days := int(age.Hours() / 24)
	if days > 0 {
		fmt.Printf("Age:      %d day(s)\n", days)
	} else {
		fmt.Printf("Age:      <1 day\n")
	}

	if entry.Status != "" {
		fmt.Printf("Status:   %s\n", entry.Status)
	}

	// Filesystem stats
	info, err := os.Stat(entry.Path)
	if err != nil {
		fmt.Printf("\n  (directory not accessible: %v)\n", err)
	} else if info.IsDir() {
		var totalSize int64
		dirCount := 0
		fileCount := 0
		var lastMod time.Time

		_ = filepath.WalkDir(entry.Path, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			if d.IsDir() {
				dirCount++
			} else {
				fileCount++
				fi, fiErr := d.Info()
				if fiErr == nil {
					totalSize += fi.Size()
					if fi.ModTime().After(lastMod) {
						lastMod = fi.ModTime()
					}
				}
			}
			return nil
		})

		fmt.Printf("Dirs:     %d\n", dirCount)
		fmt.Printf("Files:    %d\n", fileCount)
		fmt.Printf("Size:     %s\n", formatBytes(totalSize))
		if !lastMod.IsZero() {
			fmt.Printf("Modified: %s\n", lastMod.Format("2006-01-02 15:04:05"))
		}
	}

	if len(entry.Notes) > 0 {
		fmt.Println("Notes:")
		for i, n := range entry.Notes {
			fmt.Printf("  %d. %s\n", i+1, n)
		}
	}

	return nil
}

func formatBytes(b int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
