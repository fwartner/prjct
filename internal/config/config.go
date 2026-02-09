package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileTemplate represents a file to be created inside a directory.
type FileTemplate struct {
	Name    string `yaml:"name"`
	Content string `yaml:"content,omitempty"`
}

// Variable represents a user-prompted variable for template expansion.
type Variable struct {
	Name    string `yaml:"name"`
	Prompt  string `yaml:"prompt,omitempty"`
	Default string `yaml:"default,omitempty"`
}

// Directory represents a single directory node in a template tree.
type Directory struct {
	Name     string         `yaml:"name"`
	Children []Directory    `yaml:"children,omitempty"`
	Files    []FileTemplate `yaml:"files,omitempty"`
	Optional bool           `yaml:"optional,omitempty"`
	When     string         `yaml:"when,omitempty"`
}

// Template represents a project template with its directory structure.
type Template struct {
	ID          string      `yaml:"id"`
	Name        string      `yaml:"name"`
	BasePath    string      `yaml:"base_path"`
	Directories []Directory `yaml:"directories"`
	Hooks       []string    `yaml:"hooks,omitempty"`
	Variables   []Variable  `yaml:"variables,omitempty"`
	Extends     string      `yaml:"extends,omitempty"`
	Tags        []string    `yaml:"tags,omitempty"`
}

// Config is the root configuration containing all templates.
type Config struct {
	Editor    string     `yaml:"editor,omitempty"`
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
	"list":       true,
	"config":     true,
	"doctor":     true,
	"help":       true,
	"install":    true,
	"search":     true,
	"reindex":    true,
	"open":       true,
	"completion": true,
	"tree":       true,
	"path":       true,
	"recent":     true,
	"stats":      true,
	"rename":     true,
	"archive":    true,
	"export":     true,
	"import":     true,
	"init":       true,
	"diff":       true,
	"version":    true,
	"sync":       true,
	"clone":      true,
	"clean":      true,
	"note":       true,
	"info":       true,
	"validate":   true,
	"bulk":       true,
	"undo":       true,
	"readme":     true,
	"watch":      true,
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

		if len(t.Directories) == 0 && t.Extends == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".directories",
				Message: "at least one directory is required",
			})
		} else if len(t.Directories) > 0 {
			errs = append(errs, validateDirs(t.Directories, prefix+".directories", 0)...)
		}

		for j, v := range t.Variables {
			vp := fmt.Sprintf("%s.variables[%d]", prefix, j)
			if v.Name == "" {
				errs = append(errs, ValidationError{
					Field:   vp + ".name",
					Message: "variable name is required",
				})
			} else if !varNameRe.MatchString(v.Name) {
				errs = append(errs, ValidationError{
					Field:   vp + ".name",
					Message: fmt.Sprintf("variable name %q must match [a-zA-Z_][a-zA-Z0-9_]*", v.Name),
				})
			}
		}
	}

	// Validate extends references (second pass — all IDs are now known)
	for i, t := range c.Templates {
		if t.Extends == "" {
			continue
		}
		prefix := fmt.Sprintf("templates[%d]", i)
		if !seen[t.Extends] {
			errs = append(errs, ValidationError{
				Field:   prefix + ".extends",
				Message: fmt.Sprintf("extends %q references unknown template", t.Extends),
			})
		}
		if t.Extends == t.ID {
			errs = append(errs, ValidationError{
				Field:   prefix + ".extends",
				Message: "template cannot extend itself",
			})
		}
	}

	// Check for circular inheritance
	for i, t := range c.Templates {
		if t.Extends == "" {
			continue
		}
		prefix := fmt.Sprintf("templates[%d]", i)
		visited := map[string]bool{t.ID: true}
		current := t.Extends
		for current != "" {
			if visited[current] {
				errs = append(errs, ValidationError{
					Field:   prefix + ".extends",
					Message: fmt.Sprintf("circular inheritance detected involving %q", current),
				})
				break
			}
			visited[current] = true
			parent := c.FindTemplate(current)
			if parent == nil {
				break
			}
			current = parent.Extends
		}
	}

	return errs
}

const maxDepth = 20

var varNameRe = regexp.MustCompile(`^[a-zA-Z_]\w*$`)

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
		for j, f := range d.Files {
			fp := fmt.Sprintf("%s.files[%d]", p, j)
			if f.Name == "" {
				errs = append(errs, ValidationError{
					Field:   fp + ".name",
					Message: "file name is required",
				})
			}
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

// ResolveTemplate returns a fully-merged template by following the extends
// chain. The returned template is a deep copy — safe to modify.
func (c *Config) ResolveTemplate(id string) (*Template, error) {
	tmpl := c.FindTemplate(id)
	if tmpl == nil {
		return nil, fmt.Errorf("template %q not found", id)
	}

	if tmpl.Extends == "" {
		cp := *tmpl
		return &cp, nil
	}

	// Collect inheritance chain (child-first)
	chain := []*Template{tmpl}
	visited := map[string]bool{id: true}
	current := tmpl.Extends
	for current != "" {
		if visited[current] {
			return nil, fmt.Errorf("circular inheritance at %q", current)
		}
		visited[current] = true
		parent := c.FindTemplate(current)
		if parent == nil {
			return nil, fmt.Errorf("parent template %q not found", current)
		}
		chain = append(chain, parent)
		current = parent.Extends
	}

	// Merge: start from root ancestor, overlay children
	merged := Template{}
	for i := len(chain) - 1; i >= 0; i-- {
		t := chain[i]
		if t.ID != "" {
			merged.ID = t.ID
		}
		if t.Name != "" {
			merged.Name = t.Name
		}
		if t.BasePath != "" {
			merged.BasePath = t.BasePath
		}
		merged.Directories = append(merged.Directories, t.Directories...)
		merged.Hooks = append(merged.Hooks, t.Hooks...)

		// Variables: child overrides parent by name
		for _, v := range t.Variables {
			found := false
			for j, existing := range merged.Variables {
				if existing.Name == v.Name {
					merged.Variables[j] = v
					found = true
					break
				}
			}
			if !found {
				merged.Variables = append(merged.Variables, v)
			}
		}
	}

	return &merged, nil
}

// Save writes the config to disk as YAML.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// MatchesTags returns true if the template has at least one of the given tags.
// An empty filter matches everything.
func (t *Template) MatchesTags(tags []string) bool {
	if len(tags) == 0 {
		return true
	}
	tagSet := make(map[string]bool, len(t.Tags))
	for _, tag := range t.Tags {
		tagSet[strings.ToLower(tag)] = true
	}
	for _, f := range tags {
		if tagSet[strings.ToLower(f)] {
			return true
		}
	}
	return false
}

// EvalWhen evaluates a simple condition string against variables.
// Supported forms: "key == value", "key != value", "key" (truthy check).
// Returns true if the condition is met or the condition is empty.
func EvalWhen(when string, vars map[string]string) bool {
	when = strings.TrimSpace(when)
	if when == "" {
		return true
	}

	// "key != value"
	if parts := strings.SplitN(when, "!=", 2); len(parts) == 2 {
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		return vars[k] != v
	}

	// "key == value"
	if parts := strings.SplitN(when, "==", 2); len(parts) == 2 {
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		return vars[k] == v
	}

	// Truthy: variable exists and is non-empty
	return vars[when] != ""
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
