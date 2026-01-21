package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/woozymasta/dayz-stringtable/internal/poutil"
)

// TestPosCmd verifies that PosCmd generates .po files for each language.
func TestPosCmd(t *testing.T) {
	tmpDir := t.TempDir()
	// CSV format: "Language","original",...
	csvContent := `"Language","original","english","spanish"
"hello","greet","Hello","Hola"
`
	csvPath := filepath.Join(tmpDir, "in.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0o644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	poDir := filepath.Join(tmpDir, "po")
	cmd := PosCmd{Input: csvPath, Langs: "english,spanish", OutDir: poDir, Force: false}
	if err := cmd.Execute(nil); err != nil {
		t.Fatalf("PosCmd.Execute failed: %v", err)
	}

	for _, lang := range []string{"english", "spanish"} {
		p, err := poutil.ParseFile(filepath.Join(poDir, lang+".po"))
		if err != nil {
			t.Fatalf("failed to parse PO file: %v", err)
		}
		want := "Hello"
		if lang == "spanish" {
			want = "Hola"
		}
		// PO format: msgctxt = key (row[0] = "hello"), msgid = original (row[1] = "greet")
		if got := p.GetC("hello", "greet"); got != want {
			t.Errorf("%s translation = %q, want %q", lang, got, want)
		}
	}
}
