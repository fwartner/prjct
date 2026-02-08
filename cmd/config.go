package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fwartner/prjct/internal/config"
	"github.com/spf13/cobra"
)

var editConfig bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show config file location",
	Long: `Show the config file location and status.
Use --edit to open the config file in your preferred editor.

Editor resolution order:
  1. "editor" field in config YAML
  2. $VISUAL environment variable
  3. $EDITOR environment variable
  4. Platform default (open on macOS, notepad on Windows, vi on Linux)`,
	RunE: runConfig,
}

func init() {
	configCmd.Flags().BoolVarP(&editConfig, "edit", "e", false, "open config file in editor")
}

// execCommand is the function used to run external commands.
// Overridden in tests.
var execCommand = execCommandDefault

func execCommandDefault(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runConfig(cmd *cobra.Command, args []string) error {
	path := configPath
	if path == "" {
		var err error
		path, err = config.DefaultPath()
		if err != nil {
			return &ExitError{Code: ExitGeneral, Message: err.Error()}
		}
	}

	if editConfig {
		return openInEditor(path)
	}

	fmt.Printf("Config file: %s\n", path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Status:      not found")
		fmt.Println()
		fmt.Println("Run 'prjct install' to create a default config.")
	} else {
		fmt.Println("Status:      found")
	}

	return nil
}

func openInEditor(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &ExitError{
			Code:    ExitConfigNotFound,
			Message: fmt.Sprintf("config file not found: %s\nRun 'prjct install' to create a default config.", path),
		}
	}

	editor := resolveEditor(path)

	// Split editor string to handle editors with flags (e.g. "code --wait")
	parts := strings.Fields(editor)
	name := parts[0]
	editorArgs := append(parts[1:], path)

	if err := execCommand(name, editorArgs...); err != nil {
		return &ExitError{
			Code:    ExitGeneral,
			Message: fmt.Sprintf("failed to open editor %q: %v", name, err),
		}
	}

	return nil
}

// resolveEditor determines which editor to use.
// Priority: config file editor field → $VISUAL → $EDITOR → platform default.
func resolveEditor(cfgPath string) string {
	// Try config file's editor field
	cfg, err := config.Load(cfgPath)
	if err == nil && cfg.Editor != "" {
		return cfg.Editor
	}

	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	if v := os.Getenv("EDITOR"); v != "" {
		return v
	}

	return defaultEditor()
}

func defaultEditor() string {
	switch runtime.GOOS {
	case "darwin":
		return "open"
	case "windows":
		return "notepad"
	default:
		return "vi"
	}
}
