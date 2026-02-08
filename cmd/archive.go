package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

var (
	archiveDelete bool
	archiveOutput string
)

var archiveCmd = &cobra.Command{
	Use:   "archive <query>",
	Short: "Archive a project as a .tar.gz file",
	Long: `Searches the project index and creates a compressed archive of the
first matching project. Use --delete to remove the original after archiving.`,
	Args: cobra.ExactArgs(1),
	RunE: runArchive,
}

func init() {
	archiveCmd.Flags().BoolVar(&archiveDelete, "delete", false, "delete original after archiving")
	archiveCmd.Flags().StringVarP(&archiveOutput, "output", "o", "", "output file path (default: <project>.tar.gz)")
}

func runArchive(cmd *cobra.Command, args []string) error {
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

	// Verify directory exists
	info, err := os.Stat(projectPath)
	if err != nil || !info.IsDir() {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("project directory not found: %s", projectPath)}
	}

	outputPath := archiveOutput
	if outputPath == "" {
		outputPath = projectPath + ".tar.gz"
	}

	if err := createTarGz(outputPath, projectPath); err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("archive failed: %v", err)}
	}

	fmt.Printf("Archived: %s\n", outputPath)

	// Update index status
	_ = index.Update(idxPath, projectPath, func(e *index.Entry) {
		e.Status = "archived"
	})

	if archiveDelete {
		if err := os.RemoveAll(projectPath); err != nil {
			return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("archive created but failed to delete original: %v", err)}
		}
		fmt.Printf("Deleted:  %s\n", projectPath)
	}

	return nil
}

func createTarGz(outputPath, sourcePath string) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	baseName := filepath.Base(sourcePath)

	return filepath.WalkDir(sourcePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(filepath.Dir(sourcePath), path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(filepath.Join(baseName, rel[len(baseName):]))
		if header.Name == baseName+"/" {
			header.Name = baseName + "/"
		}

		// Use portable relative path
		relFromSource, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}
		if relFromSource == "." {
			header.Name = baseName + "/"
		} else {
			header.Name = filepath.ToSlash(filepath.Join(baseName, relFromSource))
			if d.IsDir() {
				header.Name += "/"
			}
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	})
}
