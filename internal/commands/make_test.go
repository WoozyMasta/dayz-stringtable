package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/woozymasta/dayz-stringtable/internal/poutil"
)

// TestMakeCmd verifies that MakeCmd correctly merges .po translations into a CSV.
func TestMakeCmd(t *testing.T) {
	// Setup temporary workspace
	tmpDir := t.TempDir()

	// Create sample input CSV
	// CSV format: "Language","original",...
	csvContent := `"Language","original"
"hello","greeting"
"bye","farewell"
`
	csvPath := filepath.Join(tmpDir, "input.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0o644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	// Prepare .po directory and files
	poDir := filepath.Join(tmpDir, "po")
	if err := os.Mkdir(poDir, 0o755); err != nil {
		t.Fatalf("failed to create po dir: %v", err)
	}

	// Build an English .po with translations
	// PO format: msgctxt = key (row[0]), msgid = original (row[1])
	po := poutil.NewFile()
	po.Language = "english"
	po.SetHeader("Language", "english")
	po.SetC("hello", "greeting", "Hello!")
	po.SetC("bye", "farewell", "Goodbye!")
	poData, err := po.MarshalText()
	if err != nil {
		t.Fatalf("failed to marshal po: %v", err)
	}
	if err := os.WriteFile(filepath.Join(poDir, "english.po"), poData, 0o644); err != nil {
		t.Fatalf("failed to write english.po: %v", err)
	}

	// Execute MakeCmd
	outputPath := filepath.Join(tmpDir, "full.csv")
	cmd := MakeCmd{
		Input:  csvPath,
		PoDir:  poDir,
		Output: outputPath,
		Force:  false,
	}
	if err := cmd.Execute(nil); err != nil {
		t.Fatalf("MakeCmd.Execute failed: %v", err)
	}

	// Read and verify output
	outData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	expected :=
		`"Language","original","english",
"hello","greeting","Hello!",
"bye","farewell","Goodbye!",
`
	if string(outData) != expected {
		t.Errorf("unexpected output:\n%s\nexpected:\n%s", string(outData), expected)
	}
}
