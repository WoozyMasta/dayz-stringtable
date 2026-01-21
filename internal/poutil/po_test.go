package poutil

import (
	"strings"
	"testing"
)

func TestNewFile(t *testing.T) {
	f := NewFile()
	if f == nil {
		t.Fatal("NewFile() returned nil")
	}
	if f.Headers == nil {
		t.Error("Headers map is nil")
	}
	if f.Entries == nil {
		t.Error("Entries slice is nil")
	}
	if len(f.Entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(f.Entries))
	}
}

func TestSetC_GetC(t *testing.T) {
	f := NewFile()

	// Set a translation
	f.SetC("context1", "msgid1", "translation1")

	// Get it back
	got := f.GetC("context1", "msgid1")
	if got != "translation1" {
		t.Errorf("GetC() = %q, want %q", got, "translation1")
	}

	// Get non-existent
	got = f.GetC("nonexistent", "msgid")
	if got != "" {
		t.Errorf("GetC(nonexistent) = %q, want empty string", got)
	}
}

func TestSetC_UpdateExisting(t *testing.T) {
	f := NewFile()

	// Set initial translation
	f.SetC("ctx", "msg", "trans1")

	// Update it
	f.SetC("ctx", "msg", "trans2")

	// Should have only one entry
	if len(f.Entries) != 1 {
		t.Errorf("Expected 1 entry after update, got %d", len(f.Entries))
	}

	got := f.GetC("ctx", "msg")
	if got != "trans2" {
		t.Errorf("GetC() after update = %q, want %q", got, "trans2")
	}
}

func TestIsTranslatedC(t *testing.T) {
	f := NewFile()

	// Empty translation
	f.SetC("ctx1", "msg1", "")
	if f.IsTranslatedC("ctx1", "msg1") {
		t.Error("IsTranslatedC() returned true for empty translation")
	}

	// Non-empty translation
	f.SetC("ctx2", "msg2", "trans")
	if !f.IsTranslatedC("ctx2", "msg2") {
		t.Error("IsTranslatedC() returned false for non-empty translation")
	}

	// Non-existent
	if f.IsTranslatedC("nonexistent", "msg") {
		t.Error("IsTranslatedC() returned true for non-existent entry")
	}
}

func TestSetHeader_GetHeader(t *testing.T) {
	f := NewFile()

	f.SetHeader("Test-Header", "test-value")

	got := f.GetHeader("Test-Header")
	if got != "test-value" {
		t.Errorf("GetHeader() = %q, want %q", got, "test-value")
	}

	// Non-existent header
	got = f.GetHeader("Nonexistent")
	if got != "" {
		t.Errorf("GetHeader(nonexistent) = %q, want empty string", got)
	}
}

func TestParseFile_Basic(t *testing.T) {
	poContent := `msgid ""
msgstr ""
"Project-Id-Version: test\\n"
"Language: ru\\n"

msgctxt "KEY1"
msgid "Original text"
msgstr "Переведенный текст"
`

	reader := strings.NewReader(poContent)
	po, err := ParseReader(reader)
	if err != nil {
		t.Fatalf("ParseReader() error = %v", err)
	}

	// Check headers
	if po.GetHeader("Project-Id-Version") != "test" {
		t.Errorf("Project-Id-Version = %q, want %q", po.GetHeader("Project-Id-Version"), "test")
	}
	if po.GetHeader("Language") != "ru" {
		t.Errorf("Language = %q, want %q", po.GetHeader("Language"), "ru")
	}
	if po.Language != "ru" {
		t.Errorf("po.Language = %q, want %q", po.Language, "ru")
	}

	// Check entry
	if len(po.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(po.Entries))
	}

	entry := po.Entries[0]
	if entry.Context != "KEY1" {
		t.Errorf("Context = %q, want %q", entry.Context, "KEY1")
	}
	if entry.MsgID != "Original text" {
		t.Errorf("MsgID = %q, want %q", entry.MsgID, "Original text")
	}
	if entry.MsgStr != "Переведенный текст" {
		t.Errorf("MsgStr = %q, want %q", entry.MsgStr, "Переведенный текст")
	}
}

func TestParseFile_WithComments(t *testing.T) {
	poContent := `msgid ""
msgstr ""

# some comment
# notranslate
msgctxt "KEY1"
msgid "Text"
msgstr ""
`

	reader := strings.NewReader(poContent)
	po, err := ParseReader(reader)
	if err != nil {
		t.Fatalf("ParseReader() error = %v", err)
	}

	if len(po.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(po.Entries))
	}

	entry := po.Entries[0]
	if len(entry.Comments) != 2 {
		t.Fatalf("Expected 2 comments, got %d", len(entry.Comments))
	}

	if !strings.Contains(entry.Comments[0], "some comment") {
		t.Errorf("Comment 1 = %q, should contain 'some comment'", entry.Comments[0])
	}
	if !strings.Contains(entry.Comments[1], "notranslate") {
		t.Errorf("Comment 2 = %q, should contain 'notranslate'", entry.Comments[1])
	}

	// Test HasNoTranslate
	if !entry.HasNoTranslate() {
		t.Error("HasNoTranslate() returned false for entry with notranslate comment")
	}
}

func TestHasNoTranslate(t *testing.T) {
	tests := []struct {
		name     string
		comments []string
		want     bool
	}{
		{
			name:     "flag comment with notranslate",
			comments: []string{"#, notranslate"},
			want:     true,
		},
		{
			name:     "regular comment with notranslate",
			comments: []string{"# notranslate"},
			want:     true,
		},
		{
			name:     "multiple flags",
			comments: []string{"#, fuzzy, notranslate"},
			want:     true,
		},
		{
			name:     "no notranslate",
			comments: []string{"# some comment"},
			want:     false,
		},
		{
			name:     "empty comments",
			comments: []string{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &Entry{
				Comments: tt.comments,
			}
			got := entry.HasNoTranslate()
			if got != tt.want {
				t.Errorf("HasNoTranslate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalText_Basic(t *testing.T) {
	f := NewFile()
	f.SetHeader("Language", "ru")
	f.SetC("KEY1", "Original", "Translated")

	data, err := f.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() error = %v", err)
	}

	// Check that it contains expected strings
	text := string(data)
	if !strings.Contains(text, `msgctxt "KEY1"`) {
		t.Error("MarshalText() output missing msgctxt")
	}
	if !strings.Contains(text, `msgid "Original"`) {
		t.Error("MarshalText() output missing msgid")
	}
	if !strings.Contains(text, `msgstr "Translated"`) {
		t.Error("MarshalText() output missing msgstr")
	}
	if !strings.Contains(text, `Language: ru`) {
		t.Error("MarshalText() output missing Language header")
	}
}

func TestMarshalText_WithComments(t *testing.T) {
	f := NewFile()
	entry := &Entry{
		Context:  "KEY1",
		MsgID:    "Text",
		MsgStr:   "Текст",
		Comments: []string{"# some comment", "# notranslate"},
	}
	f.Entries = append(f.Entries, entry)

	data, err := f.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() error = %v", err)
	}

	text := string(data)
	if !strings.Contains(text, "# some comment") {
		t.Error("MarshalText() output missing comment")
	}
	if !strings.Contains(text, "# notranslate") {
		t.Error("MarshalText() output missing notranslate comment")
	}
}

func TestMarshalText_RoundTrip(t *testing.T) {
	// Create a PO file with various features
	original := NewFile()
	original.SetHeader("Language", "ru")
	original.SetHeader("X-Generator", "test")

	entry1 := &Entry{
		Context:  "KEY1",
		MsgID:    "Text 1",
		MsgStr:   "Текст 1",
		Comments: []string{"# comment 1"},
	}
	entry2 := &Entry{
		Context: "KEY2",
		MsgID:   "Text 2",
		MsgStr:  "",
	}
	original.Entries = append(original.Entries, entry1, entry2)

	// Marshal
	data, err := original.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() error = %v", err)
	}

	// Parse back
	reader := strings.NewReader(string(data))
	parsed, err := ParseReader(reader)
	if err != nil {
		t.Fatalf("ParseReader() error = %v", err)
	}

	// Compare
	if parsed.GetHeader("Language") != original.GetHeader("Language") {
		t.Errorf("Language header: got %q, want %q", parsed.GetHeader("Language"), original.GetHeader("Language"))
	}
	if parsed.GetHeader("X-Generator") != original.GetHeader("X-Generator") {
		t.Errorf("X-Generator header: got %q, want %q", parsed.GetHeader("X-Generator"), original.GetHeader("X-Generator"))
	}

	if len(parsed.Entries) != len(original.Entries) {
		t.Fatalf("Entry count: got %d, want %d", len(parsed.Entries), len(original.Entries))
	}

	// Check first entry
	e1 := parsed.Entries[0]
	if e1.Context != entry1.Context || e1.MsgID != entry1.MsgID || e1.MsgStr != entry1.MsgStr {
		t.Errorf("Entry 1 mismatch: got Context=%q MsgID=%q MsgStr=%q, want Context=%q MsgID=%q MsgStr=%q",
			e1.Context, e1.MsgID, e1.MsgStr, entry1.Context, entry1.MsgID, entry1.MsgStr)
	}
	if len(e1.Comments) != 1 || !strings.Contains(e1.Comments[0], "comment 1") {
		t.Errorf("Entry 1 comments: got %v, want comment with 'comment 1'", e1.Comments)
	}

	// Check second entry
	e2 := parsed.Entries[1]
	if e2.Context != entry2.Context || e2.MsgID != entry2.MsgID || e2.MsgStr != entry2.MsgStr {
		t.Errorf("Entry 2 mismatch: got Context=%q MsgID=%q MsgStr=%q, want Context=%q MsgID=%q MsgStr=%q",
			e2.Context, e2.MsgID, e2.MsgStr, entry2.Context, entry2.MsgID, entry2.MsgStr)
	}
}

func TestGetEntry(t *testing.T) {
	f := NewFile()
	f.SetC("ctx1", "msg1", "trans1")
	f.SetC("ctx2", "msg2", "trans2")

	entry := f.GetEntry("ctx1", "msg1")
	if entry == nil {
		t.Fatal("GetEntry() returned nil")
	}
	if entry.MsgStr != "trans1" {
		t.Errorf("Entry.MsgStr = %q, want %q", entry.MsgStr, "trans1")
	}

	// Non-existent
	entry = f.GetEntry("nonexistent", "msg")
	if entry != nil {
		t.Error("GetEntry(nonexistent) returned non-nil")
	}
}

func TestUpdateBuildHeaders(t *testing.T) {
	f := NewFile()
	f.UpdateBuildHeaders("")

	// Should have X-Generator
	gen := f.GetHeader("X-Generator")
	if gen == "" {
		t.Error("X-Generator header is empty after UpdateBuildHeaders()")
	}
	if !strings.Contains(gen, "dayz-stringtable") {
		t.Errorf("X-Generator = %q, should contain 'dayz-stringtable'", gen)
	}

	// Project-Id-Version should not be automatically set (it's for project version, not tool version)
	proj := f.GetHeader("Project-Id-Version")
	if proj != "" {
		t.Errorf("Project-Id-Version should not be automatically set, got %q", proj)
	}

	// POT file should have POT-Creation-Date
	if f.Language == "" {
		potDate := f.GetHeader("POT-Creation-Date")
		if potDate == "" {
			t.Error("POT-Creation-Date header is empty for POT file")
		}
	}

	// PO file should have PO-Revision-Date
	f.Language = "ru"
	f.UpdateBuildHeaders("")
	poDate := f.GetHeader("PO-Revision-Date")
	if poDate == "" {
		t.Error("PO-Revision-Date header is empty for PO file")
	}
}

func TestUpdateBuildHeadersNoChange(t *testing.T) {
	// Test that PO-Revision-Date is not updated if content hasn't changed
	f := NewFile()
	f.Language = "ru"
	f.SetHeader("Language", "ru")
	f.SetC("key1", "msg1", "trans1")

	// First update - should set date
	f.UpdateBuildHeaders("")
	firstDate := f.GetHeader("PO-Revision-Date")
	if firstDate == "" {
		t.Fatal("PO-Revision-Date should be set on first update")
	}
	firstHash := f.GetHeader("X-Content-Hash")
	if firstHash == "" {
		t.Fatal("X-Content-Hash should be set after first update")
	}

	// Second update without changes - date should remain the same
	f.UpdateBuildHeaders("")
	secondDate := f.GetHeader("PO-Revision-Date")
	secondHash := f.GetHeader("X-Content-Hash")
	if secondDate != firstDate {
		t.Errorf("PO-Revision-Date should not change when content is unchanged: first=%q, second=%q", firstDate, secondDate)
	}
	if secondHash != firstHash {
		t.Errorf("X-Content-Hash should not change when content is unchanged: first=%q, second=%q", firstHash, secondHash)
	}

	// Add an entry - hash should change, and date should be updated
	f.SetC("key2", "msg2", "trans2")
	f.UpdateBuildHeaders("")
	thirdDate := f.GetHeader("PO-Revision-Date")
	thirdHash := f.GetHeader("X-Content-Hash")
	if thirdHash == firstHash {
		t.Errorf("X-Content-Hash should change when content changes: firstHash=%q, thirdHash=%q", firstHash, thirdHash)
	}
	// Date should be updated (may be same if within same minute, but hash confirms change was detected)
	if thirdHash != firstHash && thirdDate == firstDate {
		// If hash changed but date didn't, it might be same minute - that's acceptable
		// But we verify that the hash change was detected
		t.Logf("Hash changed (firstHash=%q, thirdHash=%q) but date is same (both=%q) - likely same minute, which is acceptable",
			firstHash, thirdHash, firstDate)
	}
}

func TestUpdateBuildHeaders_ProjectVersion(t *testing.T) {
	// Test that Project-Id-Version is set when provided
	f := NewFile()
	f.UpdateBuildHeaders("MyProject 1.2.3")

	proj := f.GetHeader("Project-Id-Version")
	if proj != "MyProject 1.2.3" {
		t.Errorf("Project-Id-Version = %q, want %q", proj, "MyProject 1.2.3")
	}

	// Test that Project-Id-Version is not set when empty
	f2 := NewFile()
	f2.UpdateBuildHeaders("")

	proj2 := f2.GetHeader("Project-Id-Version")
	if proj2 != "" {
		t.Errorf("Project-Id-Version = %q, want empty", proj2)
	}
}
