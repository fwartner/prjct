package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fwartner/prjct/internal/config"
	"github.com/fwartner/prjct/internal/index"
	"github.com/fwartner/prjct/internal/journal"
	"github.com/fwartner/prjct/internal/project"
	tmplpkg "github.com/fwartner/prjct/internal/template"
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
	dryRun     bool
	profile    string
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
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "preview changes without creating anything")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "config profile name (loads config.<profile>.yaml)")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(reindexCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(treeCmd)
	rootCmd.AddCommand(pathCmd)
	rootCmd.AddCommand(recentCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(renameCmd)
	rootCmd.AddCommand(archiveCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(noteCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(bulkCmd)
	rootCmd.AddCommand(undoCmd)
	rootCmd.AddCommand(readmeCmd)
	rootCmd.AddCommand(watchCmd)
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
		resolved, resolveErr := cfg.ResolveTemplate(args[0])
		if resolveErr != nil {
			ids := templateIDs(cfg)
			return &ExitError{
				Code:    ExitTemplateNotFound,
				Message: fmt.Sprintf("template %q not found. Available: %s", args[0], strings.Join(ids, ", ")),
			}
		}
		tmpl = resolved
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

	// Build variables
	vars := tmplpkg.BuiltinVars(sanitized, time.Now())

	// Prompt for custom variables in interactive mode
	if len(args) == 0 && len(tmpl.Variables) > 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for _, v := range tmpl.Variables {
			prompt := v.Prompt
			if prompt == "" {
				prompt = v.Name
			}
			if v.Default != "" {
				fmt.Printf("%s [%s]: ", prompt, v.Default)
			} else {
				fmt.Printf("%s: ", prompt)
			}
			if scanner.Scan() {
				val := strings.TrimSpace(scanner.Text())
				if val == "" {
					val = v.Default
				}
				vars[v.Name] = val
			} else {
				vars[v.Name] = v.Default
			}
		}
	} else {
		// Non-interactive: use defaults
		for _, v := range tmpl.Variables {
			vars[v.Name] = v.Default
		}
	}

	// Create directory structure
	opts := project.CreateOptions{
		Verbose:   verbose,
		DryRun:    dryRun,
		Variables: vars,
	}
	result, err := project.Create(tmpl, sanitized, opts)
	if err != nil {
		return mapCreateError(err)
	}

	// Best-effort index update — don't fail the command if indexing fails
	if !dryRun {
		if idxPath, idxErr := resolveIndexPath(); idxErr == nil {
			_ = index.Add(idxPath, index.Entry{
				Name:         sanitized,
				TemplateID:   tmpl.ID,
				TemplateName: tmpl.Name,
				Path:         result.ProjectPath,
				CreatedAt:    time.Now(),
			})
		}
		// Best-effort journal recording
		if jPath, jErr := resolveJournalPath(); jErr == nil {
			_ = journal.Append(jPath, journal.Record{
				Timestamp: time.Now(),
				Operation: journal.OpCreate,
				Details: map[string]string{
					"path":     result.ProjectPath,
					"template": tmpl.ID,
					"name":     sanitized,
				},
			})
		}
	}

	if dryRun {
		fmt.Printf("Dry run — no directories created\n")
	} else {
		fmt.Printf("Project created successfully!\n")
	}
	fmt.Printf("  Template: %s\n", result.TemplateName)
	fmt.Printf("  Name:     %s\n", sanitized)
	fmt.Printf("  Path:     %s\n", result.ProjectPath)
	fmt.Printf("  Folders:  %d\n", result.DirsCreated)
	if result.FilesCreated > 0 {
		fmt.Printf("  Files:    %d\n", result.FilesCreated)
	}

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

	// Apply profile: config.yaml → config.<profile>.yaml
	if profile != "" {
		dir := filepath.Dir(path)
		path = filepath.Join(dir, fmt.Sprintf("config.%s.yaml", profile))
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
