package template

import (
	"strings"
	"time"
)

// BuiltinVars returns the set of built-in template variables.
func BuiltinVars(projectName string, now time.Time) map[string]string {
	return map[string]string{
		"name":  projectName,
		"date":  now.Format("2006-01-02"),
		"year":  now.Format("2006"),
		"month": now.Format("01"),
		"day":   now.Format("02"),
	}
}

// Resolve replaces {key} placeholders in s with values from vars.
func Resolve(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{"+k+"}", v)
	}
	return s
}
