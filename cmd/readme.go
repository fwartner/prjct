package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fwartner/prjct/internal/config"
	"github.com/spf13/cobra"
)

var readmeOutput string

var readmeCmd = &cobra.Command{
	Use:   "readme <template-id>",
	Short: "Generate a README from a template's structure",
	Long: `Generates a Markdown document describing the template's directory
structure with file listings. Useful for documenting project layouts.`,
	Args: cobra.ExactArgs(1),
	RunE: runReadme,
}

func init() {
	readmeCmd.Flags().StringVarP(&readmeOutput, "output", "o", "", "output file (default: stdout)")
}

func runReadme(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	tmpl, err := cfg.ResolveTemplate(args[0])
	if err != nil {
		return &ExitError{Code: ExitTemplateNotFound, Message: err.Error()}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", tmpl.Name))
	b.WriteString(fmt.Sprintf("Template ID: `%s`\n\n", tmpl.ID))

	if len(tmpl.Tags) > 0 {
		b.WriteString(fmt.Sprintf("Tags: %s\n\n", strings.Join(tmpl.Tags, ", ")))
	}

	b.WriteString("## Directory Structure\n\n")
	b.WriteString("```\n")
	b.WriteString(fmt.Sprintf("%s/\n", tmpl.ID))
	writeReadmeTree(&b, tmpl.Directories, "")
	b.WriteString("```\n")

	if len(tmpl.Variables) > 0 {
		b.WriteString("\n## Variables\n\n")
		b.WriteString("| Name | Prompt | Default |\n")
		b.WriteString("|------|--------|---------|\n")
		for _, v := range tmpl.Variables {
			prompt := v.Prompt
			if prompt == "" {
				prompt = v.Name
			}
			b.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", v.Name, prompt, v.Default))
		}
	}

	if len(tmpl.Hooks) > 0 {
		b.WriteString("\n## Post-Creation Hooks\n\n")
		for _, h := range tmpl.Hooks {
			b.WriteString(fmt.Sprintf("- `%s`\n", h))
		}
	}

	output := b.String()

	if readmeOutput != "" {
		if err := os.WriteFile(readmeOutput, []byte(output), 0644); err != nil {
			return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("writing file: %v", err)}
		}
		fmt.Printf("README written to %s\n", readmeOutput)
	} else {
		fmt.Print(output)
	}

	return nil
}

func writeReadmeTree(b *strings.Builder, dirs []config.Directory, prefix string) {
	for i, d := range dirs {
		isLast := i == len(dirs)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		label := d.Name
		if d.Optional {
			label += " (optional)"
		}
		b.WriteString(fmt.Sprintf("%s%s%s/\n", prefix, connector, label))

		childPrefix := prefix + "│   "
		if isLast {
			childPrefix = prefix + "    "
		}

		for j, f := range d.Files {
			fileIsLast := j == len(d.Files)-1 && len(d.Children) == 0
			fc := "├── "
			if fileIsLast {
				fc = "└── "
			}
			b.WriteString(fmt.Sprintf("%s%s%s\n", childPrefix, fc, f.Name))
		}

		if len(d.Children) > 0 {
			writeReadmeTree(b, d.Children, childPrefix)
		}
	}
}
