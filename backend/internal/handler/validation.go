package handler

import (
	"regexp"
	"strings"
)

// namePattern allows letters, numbers, spaces, hyphens, apostrophes, and periods.
// This covers most international names while preventing injection attacks.
var namePattern = regexp.MustCompile(`^[\p{L}\p{N}\s\-'\.]+$`)

// ValidateName checks if a name contains only safe characters.
// Returns the trimmed name and true if valid, or empty string and false if invalid.
func ValidateName(name string) (string, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false
	}
	if !namePattern.MatchString(name) {
		return "", false
	}
	return name, true
}

// ValidateNameWithLength validates a name with a maximum length.
func ValidateNameWithLength(name string, maxLen int) (string, bool) {
	name, valid := ValidateName(name)
	if !valid {
		return "", false
	}
	if len(name) > maxLen {
		return "", false
	}
	return name, true
}
