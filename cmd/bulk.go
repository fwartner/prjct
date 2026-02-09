package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fwartner/prjct/internal/index"
	"github.com/fwartner/prjct/internal/project"
	tmplpkg "github.com/fwartner/prjct/internal/template"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// BulkManifest describes a set of projects to create.
type BulkManifest struct {
	Projects []BulkProject `yaml:"projects"`
}

// BulkProject is a single entry in a bulk manifest.
type BulkProject struct {
	Template string `yaml:"template"`
	Name     string `yaml:"name"`
}

var bulkCmd = &cobra.Command{
	Use:   "bulk <manifest.yaml>",
	Short: "Create multiple projects from a manifest",
	Long: `Reads a YAML manifest file and creates all listed projects. The
manifest format is:

  projects:
    - template: video
      name: "Client A Campaign"
    - template: photo
      name: "Client A Portraits"`,
	Args: cobra.ExactArgs(1),
	RunE: runBulk,
}

func runBulk(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("reading manifest: %v", err)}
	}

	var manifest BulkManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("parsing manifest: %v", err)}
	}

	if len(manifest.Projects) == 0 {
		return &ExitError{Code: ExitGeneral, Message: "manifest contains no projects"}
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	idxPath, _ := resolveIndexPath()
	created := 0
	failed := 0

	for _, bp := range manifest.Projects {
		tmpl, resolveErr := cfg.ResolveTemplate(bp.Template)
		if resolveErr != nil {
			fmt.Fprintf(os.Stderr, "  SKIP %q: template %q not found\n", bp.Name, bp.Template)
			failed++
			continue
		}

		sanitized, sanErr := project.Sanitize(bp.Name)
		if sanErr != nil {
			fmt.Fprintf(os.Stderr, "  SKIP %q: invalid name: %v\n", bp.Name, sanErr)
			failed++
			continue
		}

		vars := tmplpkg.BuiltinVars(sanitized, time.Now())
		for _, v := range tmpl.Variables {
			vars[v.Name] = v.Default
		}

		opts := project.CreateOptions{
			Verbose:   verbose,
			DryRun:    dryRun,
			Variables: vars,
		}

		result, createErr := project.Create(tmpl, sanitized, opts)
		if createErr != nil {
			fmt.Fprintf(os.Stderr, "  FAIL %q: %v\n", bp.Name, createErr)
			failed++
			continue
		}

		if !dryRun && idxPath != "" {
			_ = index.Add(idxPath, index.Entry{
				Name:         sanitized,
				TemplateID:   tmpl.ID,
				TemplateName: tmpl.Name,
				Path:         result.ProjectPath,
				CreatedAt:    time.Now(),
			})
		}

		fmt.Printf("  OK   %s (%s)\n", sanitized, result.ProjectPath)
		created++
	}

	fmt.Printf("\nCreated %d, failed %d of %d project(s)\n", created, failed, len(manifest.Projects))
	return nil
}
