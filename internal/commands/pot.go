package commands

import (
	"fmt"

	"github.com/woozymasta/dayz-stringtable/internal/utils"

	"github.com/leonelquinteros/gotext"
)

// PotCmd generates a .pot template from CSV.
// Usage: dayz-stringtable pot --input stringtable.csv --output template.pot [--force]
type PotCmd struct {
	Input  string `short:"i" long:"input" description:"CSV input file" default:"stringtable.csv"`
	Output string `short:"o" long:"output" description:"POT output file (stdout if empty)"`
	Force  bool   `short:"f" long:"force" description:"Overwrite existing file"`
}

// Execute reads CSV, finds columns by header name, and emits one .pot.
func (cmd *PotCmd) Execute(_ []string) error {
	rows, err := utils.LoadCSV(cmd.Input)
	utils.CheckErr(err)

	if len(rows) < 2 {
		return fmt.Errorf("CSV must have header and at least one data row")
	}

	po := gotext.NewPo()

	for _, row := range rows[1:] {
		if len(row) < 2 {
			continue
		}

		msgid := row[0]
		domain := row[1]
		po.SetC(domain, msgid, "")
	}

	data, err := po.MarshalText()
	utils.CheckErr(err)
	utils.WriteFile(cmd.Output, data, cmd.Force)

	return nil
}
