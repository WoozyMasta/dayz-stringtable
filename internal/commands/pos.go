package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/woozymasta/dayz-stringtable/internal/csvutil"
	"github.com/woozymasta/dayz-stringtable/internal/poutil"
)

// PosCmd generates PO files for each language from a CSV file.
//
// Usage: dayz-stringtable pos --input stringtable.csv --langs en,de,ru --outdir po/ [--force] [--project-version VERSION]
type PosCmd struct {
	Input          string `short:"i" long:"input" description:"CSV input file" default:"stringtable.csv"`
	OutDir         string `short:"d" long:"podir" description:"Directory for PO files" default:"l18n"`
	Langs          string `short:"l" long:"langs" description:"Comma-sep list of langs (default all)"`
	ProjectVersion string `short:"P" long:"project-version" description:"Set Project-Id-Version header (project name and version)"`
	Force          bool   `short:"f" long:"force" description:"Overwrite existing files"`
}

// Execute reads CSV and generates PO files for each specified language.
func (cmd *PosCmd) Execute(_ []string) error {
	rows, err := csvutil.LoadCSV(cmd.Input)
	if err != nil {
		return fmt.Errorf("failed to load CSV: %w", err)
	}

	if len(rows) < 2 {
		return fmt.Errorf("CSV must have header and at least one data row")
	}

	langs := ParseLanguages(cmd.Langs)

	// Map headers to column indices
	headers := make(map[string]int)
	for i, h := range rows[0] {
		headers[h] = i
	}

	for _, lang := range langs {
		po := poutil.NewFile()
		po.Language = lang
		po.SetHeader("Language", lang)

		// CSV format: row[0] = key, row[1] = original text
		// PO format: msgctxt = key, msgid = original text
		for _, row := range rows[1:] {
			if len(row) < 2 {
				continue
			}
			msg := ""
			if idx, ok := headers[lang]; ok && idx < len(row) {
				msg = row[idx]
			}
			po.SetC(row[0], row[1], msg)
		}

		// Update build headers after all entries are added
		po.UpdateBuildHeaders(cmd.ProjectVersion)

		data, err := po.MarshalText()
		if err != nil {
			return fmt.Errorf("failed to marshal PO file for %s: %w", lang, err)
		}

		if cmd.OutDir == "" {
			fmt.Printf("# %s.po\n", lang)
			if _, err := os.Stdout.Write(data); err != nil {
				return fmt.Errorf("failed to write to stdout: %w", err)
			}
		} else {
			path := filepath.Join(cmd.OutDir, lang+".po")
			if err := csvutil.WriteFile(path, data, cmd.Force); err != nil {
				return fmt.Errorf("failed to write %s: %w", path, err)
			}
		}
	}

	return nil
}
