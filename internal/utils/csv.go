// Package utils implements some helpers for DayZ CSV localization helper.
package utils

import (
	"encoding/csv"
	"fmt"
	"os"
)

// LoadCSV reads all records from a CSV file at given path.
func LoadCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	// check for duplicate keys (first column)
	seen := make(map[string]int)
	for i, row := range rows[1:] {
		if len(row) < 1 {
			continue
		}
		key := row[0]
		if prevLine, ok := seen[key]; ok {
			return nil, fmt.Errorf("duplicate key '%s' at row %d (already seen at row %d)", key, i+2, prevLine+2)
		}
		seen[key] = i + 1
	}

	return rows, nil
}
