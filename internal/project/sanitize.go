package project

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// illegalChars are characters not allowed in file/directory names across OSes.
const illegalChars = `<>:"/\|?*`

// windowsReserved matches Windows reserved device names (case-insensitive).
var windowsReserved = regexp.MustCompile(`(?i)^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])$`)

const maxNameBytes = 255

// Sanitize cleans a project name for safe use as a directory name.
// Spaces and Unicode are preserved. Special characters are replaced with _.
// Windows reserved names are rejected on all platforms for portability.
func Sanitize(name string) (string, error) {
	// Trim surrounding whitespace
	name = strings.TrimSpace(name)

	if name == "" {
		return "", fmt.Errorf("project name cannot be empty")
	}

	// Replace illegal filesystem characters and control characters
	var b strings.Builder
	for _, r := range name {
		if r < 0x20 || strings.ContainsRune(illegalChars, r) {
			b.WriteRune('_')
		} else {
			b.WriteRune(r)
		}
	}
	name = b.String()

	// Trim trailing dots and spaces (Windows restriction)
	name = strings.TrimRight(name, ". ")

	if name == "" {
		return "", fmt.Errorf("project name contains only invalid characters")
	}

	// Reject . and .. after trimming
	if name == "." || name == ".." {
		return "", fmt.Errorf("project name %q is not allowed", name)
	}

	// Block Windows reserved names on all platforms
	if windowsReserved.MatchString(name) {
		return "", fmt.Errorf("project name %q is a reserved system name", name)
	}

	// Also block reserved names with extensions (e.g., CON.txt)
	baseName := name
	if dot := strings.IndexByte(name, '.'); dot >= 0 {
		baseName = name[:dot]
	}
	if windowsReserved.MatchString(baseName) {
		return "", fmt.Errorf("project name %q contains a reserved system name", name)
	}

	// Enforce max byte length
	if utf8.RuneCountInString(name) > maxNameBytes {
		return "", fmt.Errorf("project name exceeds maximum length of %d characters", maxNameBytes)
	}

	return name, nil
}
