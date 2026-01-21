package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/woozymasta/dayz-stringtable/internal/csvutil"
	"github.com/woozymasta/dayz-stringtable/internal/poutil"
)

// PotCmd generates a POT template file from a CSV file.
//
// Usage: dayz-stringtable pot --input stringtable.csv --output template.pot [--force] [--project-version VERSION]
type PotCmd struct {
	Input          string `short:"i" long:"input" description:"CSV input file" default:"stringtable.csv"`
	Output         string `short:"o" long:"output" description:"POT output file (stdout if empty)"`
	ProjectVersion string `short:"P" long:"project-version" description:"Set Project-Id-Version header (project name and version)"`
	Force          bool   `short:"f" long:"force" description:"Overwrite existing file"`
}

// Execute reads CSV and generates a POT template with all original strings.
func (cmd *PotCmd) Execute(_ []string) error {
	// Compute hash of CSV file
	csvHash, err := csvutil.ComputeCSVHash(cmd.Input)
	if err != nil {
		return fmt.Errorf("failed to compute CSV hash: %w", err)
	}

	rows, err := csvutil.LoadCSV(cmd.Input)
	if err != nil {
		return fmt.Errorf("failed to load CSV: %w", err)
	}

	if len(rows) < 2 {
		return fmt.Errorf("CSV must have header and at least one data row")
	}

	// Load existing POT file if it exists to check CSV hash
	var existingPOT *poutil.File
	if cmd.Output != "" {
		if _, err := os.Stat(cmd.Output); err == nil {
			existingPOT, err = poutil.ParseFile(cmd.Output)
			if err != nil {
				// If we can't parse existing file, ignore it (will be overwritten)
				existingPOT = nil
			}
		}
	}

	po := poutil.NewFile()

	// Preserve existing headers if POT file exists
	if existingPOT != nil {
		for k, v := range existingPOT.Headers {
			po.SetHeader(k, v)
		}
	}

	// CSV format: row[0] = key, row[1] = original text
	// PO format: msgctxt = key, msgid = original text
	for _, row := range rows[1:] {
		if len(row) < 2 {
			continue
		}
		po.SetC(row[0], row[1], "")
	}

	// Check if CSV hash has changed
	csvHashChanged := true
	if existingPOT != nil {
		oldHashStr := existingPOT.GetHeader("X-CSV-Hash")
		if oldHashStr != "" {
			var oldHash uint64
			if _, err := fmt.Sscanf(oldHashStr, "%x", &oldHash); err == nil {
				csvHashChanged = (csvHash != oldHash)
			}
		}
	}

	// Update build headers after all entries are added
	po.UpdateBuildHeaders(cmd.ProjectVersion)

	// For POT files, update POT-Creation-Date only if CSV hash changed
	if csvHashChanged {
		now := time.Now().UTC().Format("2006-01-02 15:04-0700")
		po.SetHeader("POT-Creation-Date", now)
	}

	// Save CSV hash in header
	po.SetHeader("X-CSV-Hash", fmt.Sprintf("%016x", csvHash))

	data, err := po.MarshalText()
	if err != nil {
		return fmt.Errorf("failed to marshal POT: %w", err)
	}

	if err := csvutil.WriteFile(cmd.Output, data, cmd.Force); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
