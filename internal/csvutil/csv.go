// Package csvutil provides utilities for reading and writing CSV files
// used in DayZ localization workflows.
package csvutil

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cespare/xxhash/v2"
)

// LoadCSV reads all records from a CSV file at the given path.
// It validates that there are no duplicate keys in the first column.
// Returns an error if the file cannot be read or contains duplicate keys.
func LoadCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer func() { _ = f.Close() }()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	// Validate: check for duplicate keys (first column)
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

// WriteFile writes data to stdout if path is empty, or to a file otherwise.
// If force is false and the file exists, it returns an error.
// The function creates parent directories as needed.
func WriteFile(path string, data []byte, force bool) error {
	if path == "" {
		_, err := os.Stdout.Write(data)
		return err
	}

	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("file %s already exists, use --force to overwrite", path)
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// CheckErr prints an error to stderr and exits the program if err is non-nil.
// This is a convenience function for CLI commands that should exit on error.
// For library code, prefer returning errors instead.
func CheckErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

// ComputeCSVHash computes a hash of the CSV file content.
// This is used to detect if the CSV file has changed, so we can
// avoid updating POT-Creation-Date when the source CSV hasn't changed.
func ComputeCSVHash(path string) (uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer func() { _ = f.Close() }()

	h := xxhash.New()
	if _, err := io.Copy(h, f); err != nil {
		return 0, fmt.Errorf("failed to read CSV file: %w", err)
	}

	return h.Sum64(), nil
}
