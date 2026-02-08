package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRunCompletionBash(t *testing.T) {
	cmd := &cobra.Command{}
	err := runCompletion(cmd, []string{"bash"})
	if err != nil {
		t.Fatalf("runCompletion(bash) error: %v", err)
	}
}

func TestRunCompletionZsh(t *testing.T) {
	cmd := &cobra.Command{}
	err := runCompletion(cmd, []string{"zsh"})
	if err != nil {
		t.Fatalf("runCompletion(zsh) error: %v", err)
	}
}

func TestRunCompletionFish(t *testing.T) {
	cmd := &cobra.Command{}
	err := runCompletion(cmd, []string{"fish"})
	if err != nil {
		t.Fatalf("runCompletion(fish) error: %v", err)
	}
}

func TestRunCompletionPowershell(t *testing.T) {
	cmd := &cobra.Command{}
	err := runCompletion(cmd, []string{"powershell"})
	if err != nil {
		t.Fatalf("runCompletion(powershell) error: %v", err)
	}
}

func TestRunCompletionInvalid(t *testing.T) {
	cmd := &cobra.Command{}
	err := runCompletion(cmd, []string{"invalid"})
	if err == nil {
		t.Fatal("expected error for invalid shell")
	}
}
