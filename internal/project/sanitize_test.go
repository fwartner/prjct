package project

import (
	"strings"
	"testing"
)

func TestSanitizeValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple name", "My Project", "My Project"},
		{"with numbers", "Project 2026", "Project 2026"},
		{"unicode", "Projekt Ubersicht", "Projekt Ubersicht"},
		{"unicode japanese", "プロジェクト", "プロジェクト"},
		{"unicode emoji removed", "Project\x01Test", "Project_Test"},
		{"leading trailing spaces trimmed", "  My Project  ", "My Project"},
		{"hyphens and underscores", "my-project_v2", "my-project_v2"},
		{"dots in middle", "project.v2", "project.v2"},
		{"single char", "a", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sanitize(tt.input)
			if err != nil {
				t.Fatalf("Sanitize(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("Sanitize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeReplacesIllegalChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"less than", "project<1>", "project_1_"},
		{"colon", "project:name", "project_name"},
		{"double quote", `project"name"`, "project_name_"},
		{"slash", "project/name", "project_name"},
		{"backslash", "project\\name", "project_name"},
		{"pipe", "project|name", "project_name"},
		{"question mark", "project?name", "project_name"},
		{"asterisk", "project*name", "project_name"},
		{"multiple illegal", "a<b>c:d", "a_b_c_d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sanitize(tt.input)
			if err != nil {
				t.Fatalf("Sanitize(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("Sanitize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeRejectsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only spaces", "   "},
		{"only tabs", "\t\t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Sanitize(tt.input)
			if err == nil {
				t.Errorf("Sanitize(%q) expected error, got nil", tt.input)
			}
		})
	}
}

func TestSanitizeRejectsWindowsReserved(t *testing.T) {
	reserved := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
		// Case variations
		"con", "Con", "prn", "Prn", "aux", "Aux", "nul", "Nul",
		"com1", "Com1", "lpt1", "Lpt1",
	}

	for _, name := range reserved {
		t.Run(name, func(t *testing.T) {
			_, err := Sanitize(name)
			if err == nil {
				t.Errorf("Sanitize(%q) expected error for reserved name, got nil", name)
			}
			if err != nil && !strings.Contains(err.Error(), "reserved") {
				t.Errorf("Sanitize(%q) error = %v, expected 'reserved' in message", name, err)
			}
		})
	}
}

func TestSanitizeRejectsReservedWithExtension(t *testing.T) {
	tests := []string{"CON.txt", "PRN.doc", "AUX.log", "NUL.dat", "COM1.bin", "LPT1.out"}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Sanitize(name)
			if err == nil {
				t.Errorf("Sanitize(%q) expected error for reserved name with extension, got nil", name)
			}
		})
	}
}

func TestSanitizeRejectsDots(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single dot", "."},
		{"double dot", ".."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Sanitize(tt.input)
			if err == nil {
				t.Errorf("Sanitize(%q) expected error, got nil", tt.input)
			}
		})
	}
}

func TestSanitizeTrimsTrailingDotsAndSpaces(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"trailing dots", "project...", "project"},
		{"trailing spaces", "project   ", "project"},
		{"trailing mixed", "project. . .", "project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sanitize(tt.input)
			if err != nil {
				t.Fatalf("Sanitize(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("Sanitize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeRejectsControlChars(t *testing.T) {
	// Control chars should be replaced, not cause rejection
	got, err := Sanitize("project\x00name")
	if err != nil {
		t.Fatalf("Sanitize with null byte error: %v", err)
	}
	if got != "project_name" {
		t.Errorf("Sanitize with null byte = %q, want %q", got, "project_name")
	}
}

func TestSanitizeMaxLength(t *testing.T) {
	// 256 characters should be rejected
	long := strings.Repeat("a", 256)
	_, err := Sanitize(long)
	if err == nil {
		t.Error("Sanitize() expected error for name exceeding 255 chars, got nil")
	}

	// 255 characters should be accepted
	exact := strings.Repeat("a", 255)
	got, err := Sanitize(exact)
	if err != nil {
		t.Fatalf("Sanitize() unexpected error for 255-char name: %v", err)
	}
	if got != exact {
		t.Errorf("Sanitize() mangled 255-char name")
	}
}

func TestSanitizePreservesSpaces(t *testing.T) {
	got, err := Sanitize("Client Commercial 2026")
	if err != nil {
		t.Fatalf("Sanitize() error: %v", err)
	}
	if got != "Client Commercial 2026" {
		t.Errorf("Sanitize() = %q, want %q", got, "Client Commercial 2026")
	}
}

func TestSanitizePreservesUnicode(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"german", "Ubersicht Projekt"},
		{"french", "Resume du Projet"},
		{"japanese", "プロジェクト名"},
		{"chinese", "项目名称"},
		{"korean", "프로젝트"},
		{"arabic", "مشروع"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sanitize(tt.input)
			if err != nil {
				t.Fatalf("Sanitize(%q) error: %v", tt.input, err)
			}
			if got != tt.input {
				t.Errorf("Sanitize(%q) = %q, want original preserved", tt.input, got)
			}
		})
	}
}

func TestSanitizeOnlyInvalidChars(t *testing.T) {
	// A name that becomes empty after sanitization and trimming
	_, err := Sanitize("...")
	if err == nil {
		t.Error("Sanitize('...') expected error, got nil")
	}
}
