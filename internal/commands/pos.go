package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/woozymasta/dayz-stringtable/internal/utils"

	"github.com/leonelquinteros/gotext"
)

// PosCmd generates .po files for each language.
// Usage: dayz-stringtable pos --input stringtable.csv --langs en,de,ru --outdir po/ [--force]
type PosCmd struct {
	Input  string `short:"i" long:"input" description:"CSV input file" default:"stringtable.csv"`
	OutDir string `short:"d" long:"podir" description:"Directory for PO files" default:"l18n"`
	Langs  string `short:"l" long:"langs" description:"Comma-sep list of langs (default all)"`
	Force  bool   `short:"f" long:"force" description:"Overwrite existing files"`
}

// Execute reads CSV and emits PO files per lang.
func (cmd *PosCmd) Execute(_ []string) error {
	rows, err := utils.LoadCSV(cmd.Input)
	utils.CheckErr(err)

	langs := defaultLangs
	if cmd.Langs != "" {
		langs = strings.Split(cmd.Langs, ",")
	}

	// map headers to column index
	headers := make(map[string]int)
	for i, h := range rows[0] {
		headers[h] = i
	}

	for _, lang := range langs {
		po := gotext.NewPo()
		po.Language = lang

		for _, row := range rows[1:] {
			msg := ""
			if idx, ok := headers[lang]; ok && idx < len(row) {
				msg = row[idx]
			}
			po.SetC(row[1], row[0], msg)
		}

		data, err := po.MarshalText()
		utils.CheckErr(err)
		if cmd.OutDir == "" {
			fmt.Printf("# %s.po\n", lang)
			_, _ = os.Stdout.Write(data)
		} else {
			path := filepath.Join(cmd.OutDir, lang+".po")
			utils.WriteFile(path, data, cmd.Force)
		}
	}

	return nil
}
