package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanCmd_ClearsMsgstrEqualMsgid(t *testing.T) {
	tmp := t.TempDir()

	poContent := `msgid ""
msgstr ""

msgctxt "AAA"
msgid "Hello"
msgstr "Hello"

msgctxt "BBB"
msgid "Bye"
msgstr "Привет"
`
	poPath := filepath.Join(tmp, "russian.po")
	if err := os.WriteFile(poPath, []byte(poContent), 0o644); err != nil {
		t.Fatalf("write po: %v", err)
	}

	cmd := &CleanCmd{PoDir: tmp}
	if err := cmd.Execute(nil); err != nil {
		t.Fatalf("execute: %v", err)
	}

	data, err := os.ReadFile(poPath)
	if err != nil {
		t.Fatalf("read po: %v", err)
	}
	out := string(data)

	if strings.Contains(out, `msgstr "Hello"`) {
		t.Errorf("expected msgstr cleared when equal to msgid, got:\n%s", out)
	}
	if !strings.Contains(out, `msgstr "Привет"`) {
		t.Errorf("expected msgstr kept for translated entry, got:\n%s", out)
	}
}

func TestCleanCmd_RemoveUnused(t *testing.T) {
	tmp := t.TempDir()

	// Create CSV with only KEY1
	csvContent := `"Language","original"
"KEY1","Text 1"
`
	csvPath := filepath.Join(tmp, "input.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	// Create PO file with KEY1 and KEY2 (KEY2 is unused)
	poContent := `msgid ""
msgstr ""

msgctxt "KEY1"
msgid "Text 1"
msgstr "Текст 1"

msgctxt "KEY2"
msgid "Text 2"
msgstr "Текст 2"
`
	poPath := filepath.Join(tmp, "russian.po")
	if err := os.WriteFile(poPath, []byte(poContent), 0o644); err != nil {
		t.Fatalf("write po: %v", err)
	}

	cmd := &CleanCmd{
		PoDir:        tmp,
		Input:        csvPath,
		RemoveUnused: true,
	}
	if err := cmd.Execute(nil); err != nil {
		t.Fatalf("execute: %v", err)
	}

	data, err := os.ReadFile(poPath)
	if err != nil {
		t.Fatalf("read po: %v", err)
	}
	out := string(data)

	// KEY1 should remain
	if !strings.Contains(out, `msgctxt "KEY1"`) {
		t.Errorf("expected KEY1 to remain, got:\n%s", out)
	}
	if !strings.Contains(out, `msgid "Text 1"`) {
		t.Errorf("expected Text 1 to remain, got:\n%s", out)
	}

	// KEY2 should be removed
	if strings.Contains(out, `msgctxt "KEY2"`) {
		t.Errorf("expected KEY2 to be removed, got:\n%s", out)
	}
	if strings.Contains(out, `msgid "Text 2"`) {
		t.Errorf("expected Text 2 to be removed, got:\n%s", out)
	}
}

func TestCleanCmd_RemoveUnused_RequiresInput(t *testing.T) {
	cmd := &CleanCmd{
		PoDir:        "/tmp",
		RemoveUnused: true,
		// Input is not set
	}
	err := cmd.Execute(nil)
	if err == nil {
		t.Error("expected error when --remove-unused is used without --input")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("expected error about --input being required, got: %v", err)
	}
}
