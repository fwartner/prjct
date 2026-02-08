package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fwartner/prjct/internal/config"
	"github.com/fwartner/prjct/internal/project"
	"github.com/spf13/cobra"
)

// Exit codes for deterministic automation.
const (
	ExitOK               = 0
	ExitGeneral          = 1
	ExitConfigNotFound   = 2
	ExitConfigInvalid    = 3
	ExitTemplateNotFound = 4
	ExitProjectExists    = 5
	ExitPermission       = 6
	ExitCreateFailed     = 7
	ExitInvalidName      = 8
	ExitUserCancelled    = 9
)

// ExitError wraps an error with a specific exit code.
type ExitError struct {
	Code    int
	Message string
	Err     error
}

func (e *ExitError) Error() string { return e.Message }
func (e *ExitError) Unwrap() error { return e.Err }

var (
	verbose    bool
	configPath string
)

var rootCmd = &cobra.Command{
	Use:   "prjct [template] [project-name]",
	Short: "Create project directory structures from templates",
	Long: `prjct is a cross-platform CLI that creates predefined directory
structures from YAML-configured templates. Provide a template and
project name as arguments, or run interactively.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.MaximumNArgs(2),
	RunE:          runRoot,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path (overrides default)")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(installCmd)
}

// Execute runs the root command and returns an exit code.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			fmt.Fprintln(os.Stderr, "Error:", exitErr.Message)
			return exitErr.Code
		}
		fmt.Fprintln(os.Stderr, "Error:", err)
		return ExitGeneral
	}
	return ExitOK
}

func runRoot(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	var tmpl *config.Template
	var projectName string

	switch len(args) {
	case 0:
		// Interactive mode
		tmpl, projectName, err = interactive(cfg)
		if err != nil {
			return err
		}
	case 2:
		// Non-interactive mode
		tmpl = cfg.FindTemplate(args[0])
		if tmpl == nil {
			ids := templateIDs(cfg)
			return &ExitError{
				Code:    ExitTemplateNotFound,
				Message: fmt.Sprintf("template %q not found. Available: %s", args[0], strings.Join(ids, ", ")),
			}
		}
		projectName = args[1]
	default:
		return &ExitError{
			Code:    ExitGeneral,
			Message: "expected 0 or 2 arguments: prjct [template] [project-name]",
		}
	}

	// Sanitize project name
	sanitized, err := project.Sanitize(projectName)
	if err != nil {
		return &ExitError{
			Code:    ExitInvalidName,
			Message: fmt.Sprintf("invalid project name: %v", err),
		}
	}

	// Create directory structure
	result, err := project.Create(tmpl, sanitized, verbose)
	if err != nil {
		return mapCreateError(err)
	}

	fmt.Printf("Project created successfully!\n")
	fmt.Printf("  Template: %s\n", result.TemplateName)
	fmt.Printf("  Name:     %s\n", sanitized)
	fmt.Printf("  Path:     %s\n", result.ProjectPath)
	fmt.Printf("  Folders:  %d\n", result.DirsCreated)

	return nil
}

func interactive(cfg *config.Config) (*config.Template, string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	// Display template menu
	fmt.Println("Available templates:")
	fmt.Println()
	for i, t := range cfg.Templates {
		fmt.Printf("  [%d] %s (%s)\n", i+1, t.Name, t.ID)
	}
	fmt.Println()

	// Read template selection
	fmt.Print("Select template [1-" + strconv.Itoa(len(cfg.Templates)) + "]: ")
	if !scanner.Scan() {
		return nil, "", &ExitError{Code: ExitUserCancelled, Message: "cancelled"}
	}
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return nil, "", &ExitError{Code: ExitUserCancelled, Message: "cancelled"}
	}

	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(cfg.Templates) {
		return nil, "", &ExitError{
			Code:    ExitGeneral,
			Message: fmt.Sprintf("invalid selection %q: enter a number between 1 and %d", input, len(cfg.Templates)),
		}
	}
	tmpl := &cfg.Templates[num-1]

	// Read project name
	fmt.Print("Project name: ")
	if !scanner.Scan() {
		return nil, "", &ExitError{Code: ExitUserCancelled, Message: "cancelled"}
	}
	name := scanner.Text()
	if strings.TrimSpace(name) == "" {
		return nil, "", &ExitError{Code: ExitUserCancelled, Message: "cancelled: empty project name"}
	}

	return tmpl, name, nil
}

func loadConfig() (*config.Config, error) {
	path := configPath
	if path == "" {
		var err error
		path, err = config.DefaultPath()
		if err != nil {
			return nil, &ExitError{Code: ExitGeneral, Message: err.Error()}
		}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, &ExitError{
			Code:    ExitConfigNotFound,
			Message: fmt.Sprintf("config file not found: %s\nRun 'prjct install' to create a default config.", path),
		}
	}

	cfg, err := config.Load(path)
	if err != nil {
		return nil, &ExitError{
			Code:    ExitConfigInvalid,
			Message: fmt.Sprintf("invalid config: %v", err),
		}
	}

	if errs := cfg.Validate(); len(errs) > 0 {
		var msgs []string
		for _, e := range errs {
			msgs = append(msgs, e.Error())
		}
		return nil, &ExitError{
			Code:    ExitConfigInvalid,
			Message: fmt.Sprintf("config validation failed:\n  %s", strings.Join(msgs, "\n  ")),
		}
	}

	return cfg, nil
}

func mapCreateError(err error) error {
	if errors.Is(err, project.ErrProjectExists) {
		return &ExitError{Code: ExitProjectExists, Message: err.Error()}
	}
	if errors.Is(err, project.ErrPermission) {
		return &ExitError{Code: ExitPermission, Message: err.Error()}
	}
	return &ExitError{Code: ExitCreateFailed, Message: err.Error()}
}

func templateIDs(cfg *config.Config) []string {
	ids := make([]string, len(cfg.Templates))
	for i, t := range cfg.Templates {
		ids[i] = t.ID
	}
	return ids
}
