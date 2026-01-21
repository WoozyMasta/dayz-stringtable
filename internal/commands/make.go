package commands

import (
	"fmt"
	"strings"

	"github.com/woozymasta/dayz-stringtable/internal/csvutil"
	"github.com/woozymasta/dayz-stringtable/internal/poutil"
)

// MakeCmd merges PO files back into a CSV file with translations.
//
// Usage: dayz-stringtable make --input stringtable.csv --podir po/ --output full.csv [--force]
type MakeCmd struct {
	Input  string `short:"i" long:"input" description:"CSV input file" default:"stringtable.csv"`
	PoDir  string `short:"d" long:"podir" description:"Directory for PO files" default:"l18n"`
	Output string `short:"o" long:"output" description:"Merged CSV output (stdout if empty)"`
	Force  bool   `short:"f" long:"force" description:"Overwrite existing files"`
}

// Execute loads CSV and PO files, then writes a merged CSV with all translations.
func (cmd *MakeCmd) Execute(_ []string) error {
	rows, err := csvutil.LoadCSV(cmd.Input)
	if err != nil {
		return fmt.Errorf("failed to load CSV: %w", err)
	}

	poMap, err := poutil.LoadPODirectory(cmd.PoDir)
	if err != nil {
		return fmt.Errorf("failed to load PO files: %w", err)
	}

	// Determine languages in default order
	var langs []string
	for _, l := range DefaultLanguages {
		if _, ok := poMap[l]; ok {
			langs = append(langs, l)
		}
	}

	var b strings.Builder
	header := append([]string{"Language", "original"}, langs...)
	writeQuotedCSVRow(&b, header)

	// Write data rows: CSV format row[0] = key, row[1] = original text
	// PO format: msgctxt = key, msgid = original text
	for _, row := range rows[1:] {
		if len(row) < 2 {
			continue
		}
		rec := []string{row[0], row[1]}
		for _, l := range langs {
			rec = append(rec, poMap[l].GetC(row[0], row[1]))
		}
		writeQuotedCSVRow(&b, rec)
	}

	if err := csvutil.WriteFile(cmd.Output, []byte(b.String()), cmd.Force); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	return nil
}

// writeQuotedCSVRow writes a CSV row with proper quoting and escaping.
func writeQuotedCSVRow(b *strings.Builder, fields []string) {
	for _, f := range fields {
		esc := strings.ReplaceAll(f, `"`, `""`)
		b.WriteString(`"` + esc + `",`)
	}
	b.WriteByte('\n')
}
