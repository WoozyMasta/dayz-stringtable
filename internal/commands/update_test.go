package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/woozymasta/dayz-stringtable/internal/poutil"
)

func TestUpdateCmd_PreservesComments(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample CSV
	csvContent := `"Language","original"
"KEY1","Text 1"
"KEY2","Text 2"
`
	csvPath := filepath.Join(tmpDir, "input.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0o644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	// Create PO directory with existing PO file that has comments
	poDir := filepath.Join(tmpDir, "po")
	if err := os.Mkdir(poDir, 0o755); err != nil {
		t.Fatalf("failed to create po dir: %v", err)
	}

	// Create existing PO file with comments
	existingPo := poutil.NewFile()
	existingPo.Language = "russian"
	existingPo.SetHeader("Language", "russian")

	// Add entry with comments
	entry1 := &poutil.Entry{
		Context:  "KEY1",
		MsgID:    "Text 1",
		MsgStr:   "Текст 1",
		Comments: []string{"# some comment", "# notranslate"},
	}
	existingPo.Entries = append(existingPo.Entries, entry1)

	// Add entry without comments
	entry2 := &poutil.Entry{
		Context: "KEY2",
		MsgID:   "Text 2",
		MsgStr:  "Текст 2",
	}
	existingPo.Entries = append(existingPo.Entries, entry2)

	poData, err := existingPo.MarshalText()
	if err != nil {
		t.Fatalf("failed to marshal existing PO: %v", err)
	}
	if err := os.WriteFile(filepath.Join(poDir, "russian.po"), poData, 0o644); err != nil {
		t.Fatalf("failed to write existing PO: %v", err)
	}

	// Run update command
	cmd := UpdateCmd{
		Input: csvPath,
		PoDir: poDir,
		Langs: "russian",
	}
	if err := cmd.Execute(nil); err != nil {
		t.Fatalf("UpdateCmd.Execute failed: %v", err)
	}

	// Read updated PO file
	file, err := os.Open(filepath.Join(poDir, "russian.po"))
	if err != nil {
		t.Fatalf("failed to open updated PO: %v", err)
	}
	defer func() { _ = file.Close() }()

	updatedPo, err := poutil.ParseReader(file)
	if err != nil {
		t.Fatalf("failed to parse updated PO: %v", err)
	}

	// Check that comments are preserved
	entry1Updated := updatedPo.GetEntry("KEY1", "Text 1")
	if entry1Updated == nil {
		t.Fatal("Entry KEY1 not found in updated PO")
	}

	if len(entry1Updated.Comments) != 2 {
		t.Errorf("Expected 2 comments for KEY1, got %d", len(entry1Updated.Comments))
	}

	foundComment := false
	foundNoTranslate := false
	for _, comment := range entry1Updated.Comments {
		if strings.Contains(comment, "some comment") {
			foundComment = true
		}
		if strings.Contains(comment, "notranslate") {
			foundNoTranslate = true
		}
	}

	if !foundComment {
		t.Error("Comment 'some comment' not preserved")
	}
	if !foundNoTranslate {
		t.Error("Comment 'notranslate' not preserved")
	}

	// Check that translation is preserved
	if entry1Updated.MsgStr != "Текст 1" {
		t.Errorf("Translation not preserved: got %q, want %q", entry1Updated.MsgStr, "Текст 1")
	}

	// Check that entry2 still exists (without comments, which is fine)
	entry2Updated := updatedPo.GetEntry("KEY2", "Text 2")
	if entry2Updated == nil {
		t.Fatal("Entry KEY2 not found in updated PO")
	}
	if entry2Updated.MsgStr != "Текст 2" {
		t.Errorf("Translation for KEY2 not preserved: got %q, want %q", entry2Updated.MsgStr, "Текст 2")
	}
}
