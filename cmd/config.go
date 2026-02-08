package cmd

import (
	"fmt"
	"os"

	"github.com/fwartner/prjct/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show config file location",
	RunE:  runConfig,
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
