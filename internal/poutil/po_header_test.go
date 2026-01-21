package poutil

import (
	"strings"
	"testing"
)

func TestMarshalText_HeadersFormat(t *testing.T) {
	f := NewFile()
	f.SetHeader("Project-Id-Version", "")
	f.SetHeader("POT-Creation-Date", "")
	f.SetHeader("PO-Revision-Date", "")
	f.SetHeader("Last-Translator", "")
	f.SetHeader("Language-Team", "")
	f.SetHeader("Language", "ru")
	f.SetHeader("MIME-Version", "1.0")
	f.SetHeader("Content-Type", "text/plain; charset=UTF-8")
	f.SetHeader("Content-Transfer-Encoding", "8bit")
	f.SetHeader("X-Generator", "Poedit 3.6")

	data, err := f.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() error = %v", err)
	}

	text := string(data)
	lines := strings.Split(text, "\n")

	// Check that headers are written as separate quoted strings
	// Each header should be on its own line in quotes
	foundHeaders := 0
	inHeader := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == `msgstr ""` {
			inHeader = true
			continue
		}
		if inHeader && strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`) {
			// This is a header line
			foundHeaders++
			// Check format: "Key: Value\n"
			if !strings.Contains(trimmed, ":") {
				t.Errorf("Header line missing colon: %q", trimmed)
			}
		}
		if inHeader && trimmed == "" {
			// End of header section
			break
		}
	}

	if foundHeaders == 0 {
		t.Error("No header lines found in output")
	}

	// Check specific headers (in file: \n is written as \n, not \\n)
	if !strings.Contains(text, `"Project-Id-Version: \n"`) {
		t.Error("Missing Project-Id-Version header")
	}
	if !strings.Contains(text, `"Language: ru\n"`) {
		t.Error("Missing Language header")
	}
	if !strings.Contains(text, `"X-Generator: Poedit 3.6\n"`) {
		t.Error("Missing X-Generator header")
	}
}

func TestParseHeader_MultiLineFormat(t *testing.T) {
	// Test parsing headers written as separate quoted strings
	poContent := `msgid ""
msgstr ""
"Project-Id-Version: \n"
"POT-Creation-Date: \n"
"PO-Revision-Date: \n"
"Last-Translator: \n"
"Language-Team: \n"
"Language: ru\n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"X-Generator: Poedit 3.6\n"

msgctxt "KEY1"
msgid "Text"
msgstr ""
`

	reader := strings.NewReader(poContent)
	po, err := ParseReader(reader)
	if err != nil {
		t.Fatalf("ParseReader() error = %v", err)
	}

	// Check headers
	if po.GetHeader("Language") != "ru" {
		t.Errorf("Language = %q, want %q", po.GetHeader("Language"), "ru")
	}
	if po.GetHeader("MIME-Version") != "1.0" {
		t.Errorf("MIME-Version = %q, want %q", po.GetHeader("MIME-Version"), "1.0")
	}
	if po.GetHeader("X-Generator") != "Poedit 3.6" {
		t.Errorf("X-Generator = %q, want %q", po.GetHeader("X-Generator"), "Poedit 3.6")
	}
	if po.GetHeader("Project-Id-Version") != "" {
		t.Errorf("Project-Id-Version = %q, want empty", po.GetHeader("Project-Id-Version"))
	}
}
