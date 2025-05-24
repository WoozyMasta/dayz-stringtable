// Package utils implements some helpers for DayZ CSV localization helper.
package utils

import (
	"encoding/csv"
	"os"
)

// LoadCSV reads all records from a CSV file at given path.
func LoadCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return csv.NewReader(f).ReadAll()
}
