package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/woozymasta/dayz-stringtable/internal/utils"

	"github.com/leonelquinteros/gotext"
)

// UpdateCmd merges new CSV into existing .po files.
// Usage: dayz-stringtable update --input stringtable.csv --podir po/ [--langs ru,de] [--outdir updated_po/] [--force]
type UpdateCmd struct {
	Input  string `short:"i" long:"input" description:"CSV input file" default:"stringtable.csv"`
	PoDir  string `short:"d" long:"podir" description:"Directory for PO files" default:"l18n"`
	OutDir string `short:"o" long:"outdir" description:"Where to write updated PO (defaults to --podir)"`
	Langs  string `short:"l" long:"langs" description:"Comma-sep langs to update (all if empty)"`
}

// Execute reads CSV and updates each .po with new msgids.
func (cmd *UpdateCmd) Execute(_ []string) error {
	rows, err := utils.LoadCSV(cmd.Input)
	utils.CheckErr(err)

	outDir := cmd.OutDir
	if outDir == "" {
		outDir = cmd.PoDir
	}

	// load existing .po
	files, err := filepath.Glob(filepath.Join(cmd.PoDir, "*.po"))
	utils.CheckErr(err)

	poMap := map[string]*gotext.Po{}
	for _, f := range files {
		lang := strings.TrimSuffix(filepath.Base(f), ".po")
		p := gotext.NewPo()
		p.ParseFile(f)
		poMap[lang] = p
	}

	// select langs
	var langs []string
	if cmd.Langs != "" {
		langs = strings.Split(cmd.Langs, ",")
	} else {
		for l := range poMap {
			langs = append(langs, l)
		}
	}

	for _, lang := range langs {
		existing := poMap[lang]
		newPo := gotext.NewPo()
		newPo.Language = lang

		for _, row := range rows[1:] {
			if len(row) < 2 {
				continue
			}

			prev := ""
			if existing != nil {
				prev = existing.GetC(row[1], row[0])
			}
			newPo.SetC(row[1], row[0], prev)
		}

		data, err := newPo.MarshalText()
		utils.CheckErr(err)
		if err := writePOFile(outDir, lang, data); err != nil {
			return err
		}
	}

	return nil
}

// writePOFile writes .po data to disk (always overwrites).
func writePOFile(outDir, lang string, data []byte) error {
	if outDir == "" {
		return fmt.Errorf("output directory is required")
	}
	path := filepath.Join(outDir, lang+".po")
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
