package template

import (
	"testing"
	"time"
)

func TestBuiltinVars(t *testing.T) {
	now := time.Date(2026, 2, 8, 14, 30, 0, 0, time.UTC)
	vars := BuiltinVars("My Project", now)

	tests := map[string]string{
		"name":  "My Project",
		"date":  "2026-02-08",
		"year":  "2026",
		"month": "02",
		"day":   "08",
	}

	for k, want := range tests {
		if got := vars[k]; got != want {
			t.Errorf("BuiltinVars[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestResolve(t *testing.T) {
	vars := map[string]string{
		"name": "MyProject",
		"year": "2026",
	}

	tests := []struct {
		input string
		want  string
	}{
		{"{name}", "MyProject"},
		{"{name}-{year}", "MyProject-2026"},
		{"no placeholders", "no placeholders"},
		{"{unknown}", "{unknown}"},
		{"", ""},
	}

	for _, tt := range tests {
		got := Resolve(tt.input, vars)
		if got != tt.want {
			t.Errorf("Resolve(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResolveEmptyVars(t *testing.T) {
	got := Resolve("{name}", nil)
	if got != "{name}" {
		t.Errorf("Resolve with nil vars = %q, want %q", got, "{name}")
	}
}
