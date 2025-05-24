package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// CheckErr prints error to stderr and exits if err is non-nil.
func CheckErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

// WriteFile writes data to stdout if path=="" or to a file.
// It respects force flag to overwrite.
func WriteFile(path string, data []byte, force bool) {
	if path == "" {
		_, err := os.Stdout.Write(data)
		CheckErr(err)
		return
	}

	if !force {
		if _, err := os.Stat(path); err == nil {
			CheckErr(fmt.Errorf("file %s already exists, use --force to overwrite", path))
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		CheckErr(err)
	}

	CheckErr(os.WriteFile(path, data, 0o600))
}
