package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fwartner/prjct/internal/config"
	"github.com/spf13/cobra"
)

var forceInstall bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Create a default config file",
	Long:  `Creates a default configuration file with example templates (video, photo, dev).`,
	RunE:  runInstall,
}

func init() {
	installCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "overwrite existing config without prompting")
}

func runInstall(cmd *cobra.Command, args []string) error {
	path := configPath
	if path == "" {
		var err error
		path, err = config.DefaultPath()
		if err != nil {
			return &ExitError{Code: ExitGeneral, Message: err.Error()}
		}
	}

	// Check if config already exists
	if _, err := os.Stat(path); err == nil && !forceInstall {
		fmt.Printf("Config already exists: %s\n", path)
		fmt.Print("Overwrite? [y/N]: ")

		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return &ExitError{Code: ExitUserCancelled, Message: "cancelled"}
		}
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Create parent directories
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &ExitError{
			Code:    ExitPermission,
			Message: fmt.Sprintf("cannot create config directory: %v", err),
		}
	}

	// Write default config
	content := config.DefaultConfigYAML()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return &ExitError{
			Code:    ExitPermission,
			Message: fmt.Sprintf("cannot write config file: %v", err),
		}
	}

	fmt.Printf("Config created: %s\n", path)
	fmt.Println("Edit this file to customize your project templates.")

	return nil
}
