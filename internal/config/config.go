package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Directory represents a single directory node in a template tree.
type Directory struct {
	Name     string      `yaml:"name"`
	Children []Directory `yaml:"children,omitempty"`
}

// Template represents a project template with its directory structure.
type Template struct {
	ID          string      `yaml:"id"`
	Name        string      `yaml:"name"`
	BasePath    string      `yaml:"base_path"`
	Directories []Directory `yaml:"directories"`
}

// Config is the root configuration containing all templates.
type Config struct {
	Templates []Template `yaml:"templates"`
}

// ValidationError describes a single validation issue.
type ValidationError struct {
	Field   string
	Message string
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

// reservedIDs are subcommand names that cannot be used as template IDs.
var reservedIDs = map[string]bool{
	"list":    true,
	"config":  true,
	"doctor":  true,
	"help":    true,
	"install": true,
}

// Load reads and parses the config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// Validate checks the config for semantic correctness.
func (c *Config) Validate() []ValidationError {
	var errs []ValidationError

	if len(c.Templates) == 0 {
		errs = append(errs, ValidationError{
			Field:   "templates",
			Message: "no templates defined",
		})
		return errs
	}

	seen := make(map[string]bool)
	for i, t := range c.Templates {
		prefix := fmt.Sprintf("templates[%d]", i)

		if t.ID == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".id",
				Message: "id is required",
			})
		} else {
			if seen[t.ID] {
				errs = append(errs, ValidationError{
					Field:   prefix + ".id",
					Message: fmt.Sprintf("duplicate id %q", t.ID),
				})
			}
			seen[t.ID] = true

			if reservedIDs[t.ID] {
				errs = append(errs, ValidationError{
					Field:   prefix + ".id",
					Message: fmt.Sprintf("id %q conflicts with a built-in command", t.ID),
				})
			}
		}

		if t.Name == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".name",
				Message: "name is required",
			})
		}

		if t.BasePath == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".base_path",
				Message: "base_path is required",
			})
		}

		if len(t.Directories) == 0 {
			errs = append(errs, ValidationError{
				Field:   prefix + ".directories",
				Message: "at least one directory is required",
			})
		} else {
			errs = append(errs, validateDirs(t.Directories, prefix+".directories", 0)...)
		}
	}

	return errs
}

const maxDepth = 20

func validateDirs(dirs []Directory, prefix string, depth int) []ValidationError {
	var errs []ValidationError

	if depth > maxDepth {
		errs = append(errs, ValidationError{
			Field:   prefix,
			Message: fmt.Sprintf("directory nesting exceeds maximum depth of %d", maxDepth),
		})
		return errs
	}

	for i, d := range dirs {
		p := fmt.Sprintf("%s[%d]", prefix, i)
		if d.Name == "" {
			errs = append(errs, ValidationError{
				Field:   p + ".name",
				Message: "directory name is required",
			})
		}
		if len(d.Children) > 0 {
			errs = append(errs, validateDirs(d.Children, p+".children", depth+1)...)
		}
	}

	return errs
}

// FindTemplate returns the template with the given ID, or nil if not found.
func (c *Config) FindTemplate(id string) *Template {
	for i := range c.Templates {
		if c.Templates[i].ID == id {
			return &c.Templates[i]
		}
	}
	return nil
}

// ExpandPath resolves ~ to the user's home directory in a path string.
func ExpandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot expand ~: %w", err)
	}

	if path == "~" {
		return home, nil
	}

	// Handle ~/... paths
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		return filepath.Join(home, path[2:]), nil
	}

	// ~otheruser not supported
	return path, nil
}
