package commands

import (
	"path/filepath"
	"strings"

	"github.com/leonelquinteros/gotext"
	"github.com/woozymasta/dayz-stringtable/internal/utils"
)

// MakeCmd read CSV + .po files and write merged CSV
// Usage: dayz-stringtable make --input stringtable.csv --podir po/ --output full.csv [--force]
type MakeCmd struct {
	Input  string `short:"i" long:"input" description:"CSV input file" default:"stringtable.csv"`
	PoDir  string `short:"d" long:"podir" description:"Directory for PO files" default:"l18n"`
	Output string `short:"o" long:"output" description:"Merged CSV output (stdout if empty)"`
	Force  bool   `short:"f" long:"force" description:"Overwrite existing files"`
}

// Execute loads CSV and .po files then writes merged CSV.
func (cmd *MakeCmd) Execute(_ []string) error {
	rows, err := utils.LoadCSV(cmd.Input)
	utils.CheckErr(err)

	// find .po files
	files, err := filepath.Glob(filepath.Join(cmd.PoDir, "*.po"))
	utils.CheckErr(err)

	poMap := map[string]*gotext.Po{}
	for _, f := range files {
		lang := strings.TrimSuffix(filepath.Base(f), ".po")
		p := gotext.NewPo()
		p.ParseFile(f)
		poMap[lang] = p
	}

	// determine langs in default order
	var langs []string
	for _, l := range defaultLangs {
		if _, ok := poMap[l]; ok {
			langs = append(langs, l)
		}
	}

	var b strings.Builder
	// header
	header := append([]string{"Language", "original"}, langs...)
	writeQuoted := func(fields []string) {
		for _, f := range fields {
			esc := strings.ReplaceAll(f, `"`, `""`)
			b.WriteString(`"` + esc + `",`)
		}
		b.WriteByte('\n')
	}
	writeQuoted(header)

	// rows
	for _, row := range rows[1:] {
		rec := []string{row[0], row[1]}
		for _, l := range langs {
			rec = append(rec, poMap[l].GetC(row[1], row[0]))
		}
		writeQuoted(rec)
	}

	utils.WriteFile(cmd.Output, []byte(b.String()), cmd.Force)
	return nil
}
