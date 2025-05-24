package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leonelquinteros/gotext"
)

// TestMakeCmd verifies that MakeCmd correctly merges .po translations into a CSV.
func TestMakeCmd(t *testing.T) {
	// Setup temporary workspace
	tmpDir := t.TempDir()

	// Create sample input CSV
	csvContent := "msgid,domain\nhello,greeting\nbye,farewell\n"
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
	po := gotext.NewPo()
	po.Language = "english"
	po.SetC("greeting", "hello", "Hello!")
	po.SetC("farewell", "bye", "Goodbye!")
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
