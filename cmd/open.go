package cmd

import (
	"fmt"
	"runtime"

	"github.com/fwartner/prjct/internal/index"
	"github.com/spf13/cobra"
)

var openTerminal bool

var openCmd = &cobra.Command{
	Use:   "open <query>",
	Short: "Open a project in the file manager or terminal",
	Long: `Searches the project index and opens the first matching project
in your file manager. Use --terminal to open in a terminal instead.`,
	Args: cobra.ExactArgs(1),
	RunE: runOpen,
}

func init() {
	openCmd.Flags().BoolVar(&openTerminal, "terminal", false, "open in terminal instead of file manager")
}

func runOpen(cmd *cobra.Command, args []string) error {
	idxPath, err := resolveIndexPath()
	if err != nil {
		return err
	}

	idx, err := index.Load(idxPath)
	if err != nil {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("cannot load index: %v", err)}
	}

	results := index.Search(idx, args[0])
	if len(results) == 0 {
		return &ExitError{Code: ExitGeneral, Message: fmt.Sprintf("no project matching %q", args[0])}
	}

	entry := results[0]
	if openTerminal {
		return openInTerminal(entry.Path)
	}
	return openInFileManager(entry.Path)
}

func openInFileManager(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return execCommand("open", path)
	case "windows":
		return execCommand("explorer", path)
	default:
		return execCommand("xdg-open", path)
	}
}

func openInTerminal(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return execCommand("open", "-a", "Terminal", path)
	case "windows":
		return execCommand("cmd", "/c", "start", "cmd", "/k", "cd", "/d", path)
	default:
		return execCommand("x-terminal-emulator", "--working-directory", path)
	}
}
