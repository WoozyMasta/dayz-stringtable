package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestPotCmd verifies that PotCmd generates a .pot file from CSV.
func TestPotCmd(t *testing.T) {
	csvPath := filepath.Join("test_data", "stringtable.csv")
	if _, err := os.Stat(csvPath); err != nil {
		t.Fatalf("test data CSV not found at %s: %v", csvPath, err)
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "template.pot")

	cmd := &PotCmd{
		Input:  csvPath,
		Output: outPath,
		Force:  true,
	}
	if err := cmd.Execute(nil); err != nil {
		t.Fatalf("PotCmd.Execute failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read POT file: %v", err)
	}

	tests := []struct{ ctxt, id string }{
		{"STR_Error", "Error"},
		{"STR_Yes", "Yes"},
		{"STR_No", "No"},
	}

	for _, tc := range tests {
		wantCtx := []byte(`msgctxt "` + tc.ctxt + `"`)
		wantID := []byte(`msgid "` + tc.id + `"`)
		if !bytes.Contains(data, wantCtx) {
			t.Errorf("missing %s in POT output:\n%s", wantCtx, data)
		}
		if !bytes.Contains(data, wantID) {
			t.Errorf("missing %s in POT output:\n%s", wantID, data)
		}
	}
}
