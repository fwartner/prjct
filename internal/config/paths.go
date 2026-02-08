package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// DefaultPath returns the platform-specific default config file path.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, ".prjct", "config.yaml"), nil
	default: // darwin, linux
		return filepath.Join(home, ".config", "prjct", "config.yaml"), nil
	}
}
