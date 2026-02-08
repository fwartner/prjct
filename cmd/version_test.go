package cmd

import (
	"testing"
)

func TestVersionDefaults(t *testing.T) {
	if Version == "" {
		t.Error("Version should have a default value")
	}
	if Commit == "" {
		t.Error("Commit should have a default value")
	}
	if Date == "" {
		t.Error("Date should have a default value")
	}
}

func TestVersionDefaultValues(t *testing.T) {
	// When built without ldflags, these are the defaults
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Version", Version, "dev"},
		{"Commit", Commit, "none"},
		{"Date", Date, "unknown"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestVersionCmdDoesNotPanic(t *testing.T) {
	// The version command uses fmt.Printf to stdout.
	// Just ensure it doesn't panic.
	versionCmd.Run(versionCmd, nil)
}
