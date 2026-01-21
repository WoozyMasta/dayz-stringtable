// Package poutil provides utilities for reading and writing Gettext PO/POT files.
package poutil

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/woozymasta/dayz-stringtable/internal/vars"
)

// File represents a PO/POT file with its header and entries.
type File struct {
	// Headers stores all header fields from the PO file.
	// Common headers: Project-Id-Version, POT-Creation-Date, PO-Revision-Date,
	// Last-Translator, Language-Team, Language, MIME-Version, Content-Type,
	// Content-Transfer-Encoding, X-Generator, etc.
	Headers map[string]string

	// Language is the language code (e.g., "ru", "en").
	// For POT files, this is typically empty.
	Language string

	// Entries contains all translation entries (msgctxt, msgid, msgstr).
	Entries []*Entry
}

// Entry represents a single translation entry in a PO file.
type Entry struct {
	// Context is the msgctxt value (translation key/domain).
	Context string

	// MsgID is the original string (msgid).
	MsgID string

	// MsgStr is the translated string (msgstr).
	// Empty for untranslated entries.
	MsgStr string

	// Comments contains all comments associated with this entry.
	// Comments can be:
	// - "# comment" (translator comment)
	// - "#. extracted comment" (extracted comment)
	// - "#: reference" (reference comment)
	// - "#, flag" (flag comment, e.g., "#, notranslate")
	Comments []string
}

// NewFile creates a new empty PO file.
func NewFile() *File {
	return &File{
		Headers: make(map[string]string),
		Entries: []*Entry{},
	}
}

// SetC sets a translation entry with context (msgctxt).
// If an entry with the same context and msgid already exists, it updates it.
// Comments are preserved when updating existing entries.
func (f *File) SetC(context, msgid, msgstr string) {
	for _, entry := range f.Entries {
		if entry.Context == context && entry.MsgID == msgid {
			entry.MsgStr = msgstr
			// Comments are preserved - don't clear them
			return
		}
	}
	f.Entries = append(f.Entries, &Entry{
		Context: context,
		MsgID:   msgid,
		MsgStr:  msgstr,
	})
}

// GetC retrieves a translation by context and msgid.
// Returns empty string if not found or not translated.
func (f *File) GetC(context, msgid string) string {
	for _, entry := range f.Entries {
		if entry.Context == context && entry.MsgID == msgid {
			return entry.MsgStr
		}
	}
	return ""
}

// IsTranslatedC checks if an entry with given context and msgid is translated
// (has non-empty msgstr).
func (f *File) IsTranslatedC(context, msgid string) bool {
	msgstr := f.GetC(context, msgid)
	return msgstr != ""
}

// GetEntry retrieves an entry by context and msgid.
// Returns nil if not found.
func (f *File) GetEntry(context, msgid string) *Entry {
	for _, entry := range f.Entries {
		if entry.Context == context && entry.MsgID == msgid {
			return entry
		}
	}
	return nil
}

// HasNoTranslate checks if an entry has the "notranslate" flag in its comments.
func (e *Entry) HasNoTranslate() bool {
	for _, comment := range e.Comments {
		trimmed := strings.TrimSpace(comment)
		if strings.HasPrefix(trimmed, "#,") {
			// Flag comment format: "#, flag1, flag2"
			flags := strings.TrimPrefix(trimmed, "#,")
			flags = strings.TrimSpace(flags)
			if strings.Contains(flags, "notranslate") {
				return true
			}
		} else if strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, "notranslate") {
			// Also check for "# notranslate" format
			return true
		}
	}
	return false
}

// ParseFile reads and parses a PO/POT file from disk.
func ParseFile(path string) (*File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	return ParseReader(file)
}

// LoadPODirectory loads all PO files from a directory and returns a map
// of language names to parsed PO files.
// The language name is derived from the filename (without .po extension).
func LoadPODirectory(dir string) (map[string]*File, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.po"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob PO files: %w", err)
	}

	poMap := make(map[string]*File)
	for _, f := range files {
		lang := strings.TrimSuffix(filepath.Base(f), ".po")
		po, err := ParseFile(f)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", f, err)
		}
		poMap[lang] = po
	}

	return poMap, nil
}

// ParseReader parses a PO/POT file from a reader.
// The parser handles:
// - Header entry (msgid "" followed by msgstr with header fields)
// - Translation entries (msgctxt, msgid, msgstr)
// - Comments (translator, extracted, reference, and flag comments)
// - Multi-line strings (continuation lines starting with quotes)
func ParseReader(reader io.Reader) (*File, error) {
	po := NewFile()
	scanner := bufio.NewScanner(reader)

	var (
		currentEntry    *Entry
		currentSection  string // "msgctxt", "msgid", "msgstr", "header"
		headerBuffer    strings.Builder
		inHeader        = true
		headerStarted   = false
		pendingComments []string // Comments to attach to the next entry
	)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Handle comments: accumulate for next entry (skip header comments)
		if strings.HasPrefix(trimmed, "#") {
			if inHeader && !headerStarted {
				// Comments before header entry - skip
				continue
			}
			if inHeader && headerStarted {
				// Comments in header msgstr - append to buffer
				headerBuffer.WriteString(trimmed)
				headerBuffer.WriteString("\n")
				continue
			}
			// Comments for entries - accumulate for next entry
			if !inHeader {
				pendingComments = append(pendingComments, line)
			}
			continue
		}

		// Empty line - end of header or entry
		if trimmed == "" {
			if inHeader && headerStarted {
				// Parse accumulated header
				if headerBuffer.Len() > 0 {
					parseHeader(po, headerBuffer.String())
					headerBuffer.Reset()
				}
				inHeader = false
				headerStarted = false
			} else if !inHeader && currentEntry != nil && currentEntry.MsgID != "" {
				// Save completed entry
				po.Entries = append(po.Entries, currentEntry)
				currentEntry = nil
				currentSection = ""
				pendingComments = nil
			}
			continue
		}

		// Parse header entry: msgid "" followed by msgstr with header fields
		if strings.HasPrefix(trimmed, `msgid ""`) {
			headerStarted = true
			inHeader = true
			currentSection = "header"
			continue
		}

		// Parse header msgstr and continuation lines
		if inHeader {
			if strings.HasPrefix(trimmed, `msgstr "`) {
				currentSection = "header"
				value := extractQuotedValue(trimmed)
				if value != "" {
					headerBuffer.WriteString(value)
				}
				continue
			}
			if currentSection == "header" && strings.HasPrefix(trimmed, `"`) {
				// Continuation line in header (each quoted string is a separate header line)
				value := extractQuotedValue(trimmed)
				if value != "" {
					parseHeaderLine(po, value)
				}
				continue
			}
		}

		// Parse regular translation entries
		if strings.HasPrefix(trimmed, "msgctxt ") {
			// Save previous entry if exists
			if currentEntry != nil && currentEntry.MsgID != "" {
				po.Entries = append(po.Entries, currentEntry)
			}
			// Start new entry with pending comments
			commentsCopy := make([]string, len(pendingComments))
			copy(commentsCopy, pendingComments)
			currentEntry = &Entry{Comments: commentsCopy}
			pendingComments = nil
			currentSection = "msgctxt"
			currentEntry.Context = extractQuotedValue(trimmed)
			inHeader = false
		} else if strings.HasPrefix(trimmed, "msgid ") {
			value := extractQuotedValue(trimmed)
			// Skip header entry (empty msgid)
			if value == "" && inHeader {
				continue
			}
			// Create entry if needed
			if currentEntry == nil {
				commentsCopy := make([]string, len(pendingComments))
				copy(commentsCopy, pendingComments)
				currentEntry = &Entry{Comments: commentsCopy}
				pendingComments = nil
			}
			currentSection = "msgid"
			currentEntry.MsgID = value
			inHeader = false
		} else if strings.HasPrefix(trimmed, "msgstr ") {
			if currentEntry == nil {
				commentsCopy := make([]string, len(pendingComments))
				copy(commentsCopy, pendingComments)
				currentEntry = &Entry{Comments: commentsCopy}
				pendingComments = nil
			}
			currentSection = "msgstr"
			currentEntry.MsgStr = extractQuotedValue(trimmed)
			inHeader = false
		} else if strings.HasPrefix(trimmed, `"`) && !inHeader {
			// Continuation line for multi-line strings
			value := extractQuotedValue(trimmed)
			if currentEntry != nil {
				switch currentSection {
				case "msgctxt":
					currentEntry.Context += value
				case "msgid":
					currentEntry.MsgID += value
				case "msgstr":
					currentEntry.MsgStr += value
				}
			}
		}
	}

	// Save last entry
	if currentEntry != nil && currentEntry.MsgID != "" {
		po.Entries = append(po.Entries, currentEntry)
	}

	// Parse header if still in buffer
	if inHeader && headerBuffer.Len() > 0 {
		parseHeader(po, headerBuffer.String())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Extract language from headers
	if lang, ok := po.Headers["Language"]; ok {
		po.Language = lang
	}

	return po, nil
}

// extractQuotedValue extracts the quoted string value from a PO line.
// It handles escape sequences (\n, \t, \r, \\, \") and finds the matching end quote.
func extractQuotedValue(line string) string {
	// Find first quote
	start := strings.Index(line, `"`)
	if start == -1 {
		return ""
	}

	// Find matching end quote (handling escaped quotes)
	var result strings.Builder
	escaped := false
	for i := start + 1; i < len(line); i++ {
		if escaped {
			// Handle escape sequences
			switch line[i] {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '\\':
				result.WriteByte('\\')
			case '"':
				result.WriteByte('"')
			default:
				// Unknown escape, write as-is
				result.WriteByte('\\')
				result.WriteByte(line[i])
			}
			escaped = false
		} else if line[i] == '\\' {
			escaped = true
		} else if line[i] == '"' {
			break
		} else {
			result.WriteByte(line[i])
		}
	}

	return result.String()
}

// parseHeader parses the header string (from msgstr "" entry) and populates Headers.
// Header format: "Key: Value\n" where \n is escaped as \\n.
func parseHeader(po *File, headerStr string) {
	lines := strings.Split(headerStr, "\\n")
	for _, line := range lines {
		parseHeaderLine(po, line)
	}
}

// parseHeaderLine parses a single header line in format "Key: Value".
// Removes trailing \n if present and extracts key-value pair.
func parseHeaderLine(po *File, line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	// Header format: "Key: Value\n" or "Key: Value"
	// Remove trailing \n if present
	line = strings.TrimSuffix(line, "\\n")
	idx := strings.Index(line, ":")
	if idx > 0 {
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		po.Headers[key] = value
	}
}

// MarshalText serializes the PO file to text format.
func (f *File) MarshalText() ([]byte, error) {
	var b strings.Builder

	// Write header entry
	b.WriteString(`msgid ""` + "\n")
	b.WriteString(`msgstr ""` + "\n")

	// Write headers as separate quoted strings (one per line)
	standardHeaders := []string{
		"Project-Id-Version",
		"POT-Creation-Date",
		"PO-Revision-Date",
		"Last-Translator",
		"Language-Team",
		"Language",
		"MIME-Version",
		"Content-Type",
		"Content-Transfer-Encoding",
		"X-Generator",
	}

	written := make(map[string]bool)
	for _, key := range standardHeaders {
		if value, ok := f.Headers[key]; ok {
			// Write each header as a separate quoted string line
			// Format: "Key: Value\n" where \n is escaped as \\n
			headerLine := fmt.Sprintf("%s: %s\n", key, value)
			b.WriteString(`"`)
			// Escape the header line
			for _, r := range headerLine {
				switch r {
				case '\\':
					b.WriteString(`\\`)
				case '"':
					b.WriteString(`\"`)
				case '\n':
					b.WriteString(`\n`)
				default:
					b.WriteRune(r)
				}
			}
			b.WriteString(`"` + "\n")
			written[key] = true
		}
	}

	// Write remaining headers
	for key, value := range f.Headers {
		if !written[key] {
			headerLine := fmt.Sprintf("%s: %s\n", key, value)
			b.WriteString(`"`)
			// Escape the header line
			for _, r := range headerLine {
				switch r {
				case '\\':
					b.WriteString(`\\`)
				case '"':
					b.WriteString(`\"`)
				case '\n':
					b.WriteString(`\n`)
				default:
					b.WriteRune(r)
				}
			}
			b.WriteString(`"` + "\n")
		}
	}

	b.WriteString("\n")

	// Write entries
	for _, entry := range f.Entries {
		// Write comments
		for _, comment := range entry.Comments {
			b.WriteString(comment)
			b.WriteString("\n")
		}

		// Write msgctxt if present
		if entry.Context != "" {
			b.WriteString("msgctxt ")
			writeQuotedString(&b, entry.Context)
			b.WriteString("\n")
		}

		// Write msgid
		b.WriteString("msgid ")
		writeQuotedString(&b, entry.MsgID)
		b.WriteString("\n")

		// Write msgstr
		b.WriteString("msgstr ")
		writeQuotedString(&b, entry.MsgStr)
		b.WriteString("\n")

		b.WriteString("\n")
	}

	return []byte(b.String()), nil
}

// writeQuotedString writes a string in PO format (with proper escaping).
// Multi-line strings are written as continuation lines.
func writeQuotedString(b *strings.Builder, s string) {
	if s == "" {
		b.WriteString(`""`)
		return
	}

	// Check if string contains newlines
	if strings.Contains(s, "\n") {
		// Write as multi-line with continuation
		lines := strings.Split(s, "\n")
		for i, line := range lines {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(`"`)
			// Escape special characters
			for _, r := range line {
				switch r {
				case '\\':
					b.WriteString(`\\`)
				case '"':
					b.WriteString(`\"`)
				case '\t':
					b.WriteString(`\t`)
				default:
					b.WriteRune(r)
				}
			}
			// Add \n escape at end of line (except last)
			if i < len(lines)-1 {
				b.WriteString(`\n`)
			}
			b.WriteString(`"`)
		}
	} else {
		// Single line
		b.WriteString(`"`)
		// Escape special characters
		for _, r := range s {
			switch r {
			case '\\':
				b.WriteString(`\\`)
			case '"':
				b.WriteString(`\"`)
			case '\t':
				b.WriteString(`\t`)
			default:
				b.WriteRune(r)
			}
		}
		b.WriteString(`"`)
	}
}

// SetHeader sets a header field value.
func (f *File) SetHeader(key, value string) {
	if f.Headers == nil {
		f.Headers = make(map[string]string)
	}
	f.Headers[key] = value
}

// GetHeader retrieves a header field value.
func (f *File) GetHeader(key string) string {
	return f.Headers[key]
}

// computeContentHash computes a hash of the PO file content (excluding date headers).
// This is used to detect if the file content has actually changed, so we can
// avoid updating PO-Revision-Date when only dates have changed.
func (f *File) computeContentHash() uint64 {
	h := xxhash.New()

	// Hash language
	_, _ = io.WriteString(h, f.Language)
	_, _ = io.WriteString(h, "\n")

	// Hash all headers except date-related ones
	excludedHeaders := map[string]bool{
		"PO-Revision-Date":  true,
		"POT-Creation-Date": true,
		"X-Content-Hash":    true, // Exclude hash itself
	}

	// Sort headers for consistent hashing
	headerKeys := make([]string, 0, len(f.Headers))
	for key := range f.Headers {
		if !excludedHeaders[key] {
			headerKeys = append(headerKeys, key)
		}
	}
	// Simple sort for consistency
	for i := 0; i < len(headerKeys)-1; i++ {
		for j := i + 1; j < len(headerKeys); j++ {
			if headerKeys[i] > headerKeys[j] {
				headerKeys[i], headerKeys[j] = headerKeys[j], headerKeys[i]
			}
		}
	}

	for _, key := range headerKeys {
		_, _ = io.WriteString(h, key)
		_, _ = io.WriteString(h, ":")
		_, _ = io.WriteString(h, f.Headers[key])
		_, _ = io.WriteString(h, "\n")
	}

	// Hash all entries
	for _, entry := range f.Entries {
		_, _ = io.WriteString(h, entry.Context)
		_, _ = io.WriteString(h, "\n")
		_, _ = io.WriteString(h, entry.MsgID)
		_, _ = io.WriteString(h, "\n")
		_, _ = io.WriteString(h, entry.MsgStr)
		_, _ = io.WriteString(h, "\n")
		// Hash comments for consistency
		for _, comment := range entry.Comments {
			_, _ = io.WriteString(h, comment)
			_, _ = io.WriteString(h, "\n")
		}
		_, _ = io.WriteString(h, "\n")
	}

	return h.Sum64()
}

// UpdateBuildHeaders updates build-related headers from vars package.
// This should be called before saving PO files to ensure they have current build info.
// The PO-Revision-Date is only updated if the file content has actually changed.
// If projectVersion is not empty, it will be set as Project-Id-Version.
func (f *File) UpdateBuildHeaders(projectVersion string) {
	buildInfo := vars.Info()

	// Set or update X-Generator
	f.SetHeader("X-Generator", fmt.Sprintf("dayz-stringtable %s", buildInfo.Version))

	// Set Project-Id-Version if provided
	// Note: Project-Id-Version should contain the project name and version being translated,
	// not the tool version. It should be set manually by the user if needed.
	if projectVersion != "" {
		f.SetHeader("Project-Id-Version", projectVersion)
	}

	// Compute hash of current content (before updating dates)
	newHash := f.computeContentHash()
	oldHashStr := f.GetHeader("X-Content-Hash")

	// Check if content has changed
	contentChanged := true
	if oldHashStr != "" {
		var oldHash uint64
		if _, err := fmt.Sscanf(oldHashStr, "%x", &oldHash); err == nil {
			contentChanged = (newHash != oldHash)
		}
	}

	// Update POT-Creation-Date or PO-Revision-Date only if content changed
	if contentChanged {
		now := time.Now().UTC().Format("2006-01-02 15:04-0700")
		if f.Language == "" {
			// POT file - only update if not managed by CSV hash
			// (POT files with X-CSV-Hash are managed by pot.go command)
			if f.GetHeader("X-CSV-Hash") == "" {
				f.SetHeader("POT-Creation-Date", now)
			}
		} else {
			// PO file
			f.SetHeader("PO-Revision-Date", now)
		}
		// Save new hash
		f.SetHeader("X-Content-Hash", fmt.Sprintf("%016x", newHash))
	}
}
