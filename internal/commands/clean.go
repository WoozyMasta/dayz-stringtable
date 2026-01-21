package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/woozymasta/dayz-stringtable/internal/csvutil"
	"github.com/woozymasta/dayz-stringtable/internal/poutil"
)

// CleanCmd clears msgstr entries that are identical to msgid in PO files.
// This is useful for cleaning up machine-translated or copied entries.
//
// Usage: dayz-stringtable clean --podir l18n [--clear-only] [--input csv] [--remove-unused]
type CleanCmd struct {
	PoDir        string   `short:"d" long:"podir" description:"Directory for PO files" default:"l18n"`
	Input        string   `short:"i" long:"input" description:"CSV input file (required for --remove-unused)"`
	Langs        []string `short:"l" long:"lang" description:"Filter by languages (comma-separated or repeatable)"`
	ClearOnly    bool     `short:"c" long:"clear-only" description:"Don't add notranslate comment, just clear msgstr"`
	RemoveUnused bool     `short:"u" long:"remove-unused" description:"Remove entries not present in CSV file"`
}

// Execute processes all PO files in the directory and clears msgstr entries that match msgid.
func (cmd *CleanCmd) Execute(_ []string) error {
	if cmd.RemoveUnused && cmd.Input == "" {
		return fmt.Errorf("--input is required when using --remove-unused")
	}

	// Load CSV if needed for removing unused entries
	var validKeys map[string]bool
	if cmd.RemoveUnused {
		rows, err := csvutil.LoadCSV(cmd.Input)
		if err != nil {
			return fmt.Errorf("failed to load CSV: %w", err)
		}
		if len(rows) < 2 {
			return fmt.Errorf("CSV must have header and at least one data row")
		}

		// Build set of valid keys (context|msgid pairs) from CSV
		validKeys = make(map[string]bool)
		for _, row := range rows[1:] {
			if len(row) < 2 {
				continue
			}
			// CSV format: row[0] = key, row[1] = original text
			// PO format: msgctxt = key, msgid = original text
			// Use separator to create unique composite key
			key := row[0] + "|" + row[1]
			validKeys[key] = true
		}
	}

	files, err := filepath.Glob(filepath.Join(cmd.PoDir, "*.po"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no PO files found in %s", cmd.PoDir)
	}

	perLangCleaned := make(map[string]int)
	perLangRemoved := make(map[string]int)
	var totalCleaned, totalRemoved int

	for _, path := range files {
		lang := ExtractLanguageName(path)

		if len(cmd.Langs) > 0 && !ContainsLanguage(cmd.Langs, lang) {
			continue
		}

		cleaned, removed, err := cmd.cleanPOFile(path, validKeys)
		if err != nil {
			return fmt.Errorf("clean %s: %w", path, err)
		}
		totalCleaned += cleaned
		totalRemoved += removed
		perLangCleaned[lang] += cleaned
		perLangRemoved[lang] += removed
	}

	// summary per language (only if something was cleaned or removed)
	if totalCleaned > 0 || totalRemoved > 0 {
		for _, lang := range orderLangs(perLangCleaned, perLangRemoved) {
			cleaned := perLangCleaned[lang]
			removed := perLangRemoved[lang]
			if cleaned > 0 && removed > 0 {
				fmt.Printf("lang %s: %d cleaned, %d removed\n", lang, cleaned, removed)
			} else if cleaned > 0 {
				fmt.Printf("lang %s: %d cleaned\n", lang, cleaned)
			} else if removed > 0 {
				fmt.Printf("lang %s: %d removed\n", lang, removed)
			}
		}
	}
	return nil
}

// cleanPOFile processes a single PO file, clearing duplicate msgstr and optionally removing unused entries.
// Returns the number of cleaned and removed entries.
func (cmd *CleanCmd) cleanPOFile(path string, validKeys map[string]bool) (cleaned int, removed int, err error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = file.Close() }()

	po, err := poutil.ParseReader(file)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse PO file: %w", err)
	}

	// First pass: clear msgstr entries that duplicate msgid
	for _, entry := range po.Entries {
		if entry.MsgStr != "" && entry.MsgStr == entry.MsgID {
			entry.MsgStr = ""
			cleaned++

			// Add notranslate comment unless --clear-only is set
			if !cmd.ClearOnly {
				hasNoTranslate := false
				for _, comment := range entry.Comments {
					if strings.Contains(comment, "notranslate") {
						hasNoTranslate = true
						break
					}
				}
				if !hasNoTranslate {
					// Prepend comment so it appears before the entry
					entry.Comments = append([]string{"# notranslate"}, entry.Comments...)
				}
			}
		}
	}

	// Second pass: remove unused entries if --remove-unused is set
	if cmd.RemoveUnused && validKeys != nil {
		var filteredEntries []*poutil.Entry
		for _, entry := range po.Entries {
			key := entry.Context + "|" + entry.MsgID
			if validKeys[key] {
				filteredEntries = append(filteredEntries, entry)
			} else {
				removed++
			}
		}
		po.Entries = filteredEntries
	}

	if cleaned == 0 && removed == 0 {
		return 0, 0, nil
	}

	// Update build headers after modifications
	po.UpdateBuildHeaders("")

	// Write back
	data, err := po.MarshalText()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to marshal PO file: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return 0, 0, err
	}

	return cleaned, removed, nil
}

// orderLangs returns languages in sorted order from the given maps.
func orderLangs(cleaned, removed map[string]int) []string {
	langSet := make(map[string]bool)
	for l := range cleaned {
		langSet[l] = true
	}
	for l := range removed {
		langSet[l] = true
	}
	var langs []string
	for l := range langSet {
		langs = append(langs, l)
	}
	sort.Strings(langs)
	return langs
}
