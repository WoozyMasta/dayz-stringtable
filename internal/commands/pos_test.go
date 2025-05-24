package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leonelquinteros/gotext"
)

// TestPosCmd verifies that PosCmd generates .po files for each language.
func TestPosCmd(t *testing.T) {
	tmpDir := t.TempDir()
	csvContent := "msgid,domain,english,spanish\nhello,greet,Hello,Hola\n"
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
		p := gotext.NewPo()
		p.ParseFile(filepath.Join(poDir, lang+".po"))
		want := "Hello"
		if lang == "spanish" {
			want = "Hola"
		}
		if got := p.GetC("greet", "hello"); got != want {
			t.Errorf("%s translation = %q, want %q", lang, got, want)
		}
	}
}
