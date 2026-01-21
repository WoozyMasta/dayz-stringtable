// Package commands implements CLI commands for DayZ CSV localization helper.
// It provides commands for generating POT templates, creating PO files,
// merging translations, updating existing files, showing statistics, and cleaning entries.
package commands

import (
	"path/filepath"
	"strings"
)

// DefaultLanguages is the list of default languages used when no specific languages are provided.
var DefaultLanguages = []string{
	"english",
	"czech",
	"german",
	"russian",
	"polish",
	"hungarian",
	"italian",
	"spanish",
	"french",
	"chinese",
	"japanese",
	"portuguese",
	"chinesesimp",
}

// ParseLanguages parses a comma-separated string of languages into a slice.
func ParseLanguages(langsStr string) []string {
	if langsStr == "" {
		return DefaultLanguages
	}
	return strings.Split(langsStr, ",")
}

// ExtractLanguageName extracts the language name from a PO file path.
// It returns the base filename without the .po extension.
func ExtractLanguageName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), ".po")
}

// ContainsLanguage checks if a language is in the given list.
func ContainsLanguage(list []string, lang string) bool {
	for _, item := range list {
		if item == lang {
			return true
		}
	}
	return false
}
