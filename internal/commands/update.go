package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/woozymasta/dayz-stringtable/internal/csvutil"
	"github.com/woozymasta/dayz-stringtable/internal/poutil"
)

// UpdateCmd merges new strings from CSV into existing PO files.
//
// Usage: dayz-stringtable update --input stringtable.csv --podir po/ [--langs ru,de] [--outdir updated_po/] [--project-version VERSION]
type UpdateCmd struct {
	Input          string `short:"i" long:"input" description:"CSV input file" default:"stringtable.csv"`
	PoDir          string `short:"d" long:"podir" description:"Directory for PO files" default:"l18n"`
	OutDir         string `short:"o" long:"outdir" description:"Where to write updated PO (defaults to --podir)"`
	Langs          string `short:"l" long:"langs" description:"Comma-sep langs to update (all if empty)"`
	ProjectVersion string `short:"P" long:"project-version" description:"Set Project-Id-Version header (project name and version)"`
}

// Execute reads CSV and updates each PO file with new entries, preserving existing translations.
func (cmd *UpdateCmd) Execute(_ []string) error {
	rows, err := csvutil.LoadCSV(cmd.Input)
	if err != nil {
		return fmt.Errorf("failed to load CSV: %w", err)
	}

	outDir := cmd.OutDir
	if outDir == "" {
		outDir = cmd.PoDir
	}

	poMap, err := poutil.LoadPODirectory(cmd.PoDir)
	if err != nil {
		return fmt.Errorf("failed to load PO files: %w", err)
	}

	// Select languages to update
	var langs []string
	if cmd.Langs != "" {
		langs = ParseLanguages(cmd.Langs)
		// Filter to only languages that exist
		filtered := make([]string, 0, len(langs))
		for _, lang := range langs {
			if _, ok := poMap[lang]; ok {
				filtered = append(filtered, lang)
			}
		}
		langs = filtered
	} else {
		for l := range poMap {
			langs = append(langs, l)
		}
	}

	for _, lang := range langs {
		existing := poMap[lang]
		newPo := poutil.NewFile()
		newPo.Language = lang

		// Preserve existing headers
		if existing != nil {
			for k, v := range existing.Headers {
				newPo.SetHeader(k, v)
			}
		}
		newPo.SetHeader("Language", lang)

		// CSV format: row[0] = key, row[1] = original text
		// PO format: msgctxt = key, msgid = original text
		for _, row := range rows[1:] {
			if len(row) < 2 {
				continue
			}

			key := row[0]
			original := row[1]

			// Get existing entry to preserve translation and comments
			var existingEntry *poutil.Entry
			if existing != nil {
				existingEntry = existing.GetEntry(key, original)
			}

			prevMsgStr := ""
			if existingEntry != nil {
				prevMsgStr = existingEntry.MsgStr
			}

			// Set the entry (updates if exists, creates new otherwise)
			newPo.SetC(key, original, prevMsgStr)

			// Preserve comments from existing entry
			if existingEntry != nil && len(existingEntry.Comments) > 0 {
				newEntry := newPo.GetEntry(key, original)
				if newEntry != nil {
					newEntry.Comments = existingEntry.Comments
				}
			}
		}

		// Update build headers after all entries are added
		newPo.UpdateBuildHeaders(cmd.ProjectVersion)

		data, err := newPo.MarshalText()
		if err != nil {
			return fmt.Errorf("failed to marshal PO file for %s: %w", lang, err)
		}

		if err := writePOFile(outDir, lang, data); err != nil {
			return fmt.Errorf("failed to write PO file for %s: %w", lang, err)
		}
	}

	return nil
}

// writePOFile writes PO file data to disk, creating parent directories as needed.
func writePOFile(outDir, lang string, data []byte) error {
	if outDir == "" {
		return fmt.Errorf("output directory is required")
	}
	path := filepath.Join(outDir, lang+".po")
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
