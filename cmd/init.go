package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fwartner/prjct/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	initID     string
	initName   string
	initOutput string
)

var initCmd = &cobra.Command{
	Use:   "init <path>",
	Short: "Generate a template from an existing directory",
	Long: `Scans an existing directory structure and generates a YAML template
configuration that can be used to recreate it.`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVar(&initID, "id", "", "template ID (default: directory name)")
	initCmd.Flags().StringVar(&initName, "name", "", "template display name (default: directory name)")
	initCmd.Flags().StringVarP(&initOutput, "output", "o", "", "output file (default: stdout)")
}

func runInit(cmd *cobra.Command, args []string) error {
	rootPath := args[0]
	info, err := os.Stat(rootPath)
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("cannot access path: %v", err)}
	}
	if !info.IsDir() {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("not a directory: %s", rootPath)}
	}

	dirName := filepath.Base(rootPath)
	id := initID
	if id == "" {
		id = dirName
	}
	name := initName
	if name == "" {
		name = dirName
	}

	dirs, err := scanDir(rootPath)
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("scanning directory: %v", err)}
	}

	tmpl := config.Template{
		ID:          id,
		Name:        name,
		BasePath:    filepath.Dir(rootPath),
		Directories: dirs,
	}

	wrapped := &config.Config{Templates: []config.Template{tmpl}}
	data, err := yaml.Marshal(wrapped)
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("marshaling: %v", err)}
	}

	if initOutput != "" {
		if err := os.WriteFile(initOutput, data, 0644); err != nil {
			return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("writing file: %v", err)}
		}
		fmt.Printf("Template written to %s\n", initOutput)
	} else {
		fmt.Print(string(data))
	}

	return nil
}

func scanDir(path string) ([]config.Directory, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []config.Directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip hidden directories
		if entry.Name()[0] == '.' {
			continue
		}

		d := config.Directory{Name: entry.Name()}
		children, err := scanDir(filepath.Join(path, entry.Name()))
		if err != nil {
			// Skip directories we cannot read (e.g. permission denied)
			if os.IsPermission(err) {
				dirs = append(dirs, d)
				continue
			}
			return nil, err
		}
		if len(children) > 0 {
			d.Children = children
		}
		dirs = append(dirs, d)
	}
	return dirs, nil
}
