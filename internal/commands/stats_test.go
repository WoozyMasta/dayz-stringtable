package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/woozymasta/dayz-stringtable/internal/poutil"
)

// TestStatsCmd_BasicOutput verifies basic statistics output
func TestStatsCmd_BasicOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample CSV
	csvContent := `"Language","original","english","russian"
"STR_Yes","Yes","Yes","Да"
"STR_No","No","No","Нет"
"STR_Error","Error","Error",""
"STR_Hello","Hello","Hello",""
`
	csvPath := filepath.Join(tmpDir, "input.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0o644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	// Create PO directory and files
	poDir := filepath.Join(tmpDir, "po")
	if err := os.Mkdir(poDir, 0o755); err != nil {
		t.Fatalf("failed to create po dir: %v", err)
	}

	// Create russian.po with some translations
	// PO format: msgctxt = key (row[0] = "STR_Yes"), msgid = original (row[1] = "Yes")
	ruPo := poutil.NewFile()
	ruPo.Language = "russian"
	ruPo.SetHeader("Language", "russian")
	ruPo.SetC("STR_Yes", "Yes", "Да")
	ruPo.SetC("STR_No", "No", "Нет")
	// STR_Error and STR_Hello are not translated
	ruData, err := ruPo.MarshalText()
	if err != nil {
		t.Fatalf("failed to marshal russian.po: %v", err)
	}
	if err := os.WriteFile(filepath.Join(poDir, "russian.po"), ruData, 0o644); err != nil {
		t.Fatalf("failed to write russian.po: %v", err)
	}

	// Create english.po with all translations
	enPo := poutil.NewFile()
	enPo.Language = "english"
	enPo.SetHeader("Language", "english")
	enPo.SetC("STR_Yes", "Yes", "Yes")
	enPo.SetC("STR_No", "No", "No")
	enPo.SetC("STR_Error", "Error", "Error")
	enPo.SetC("STR_Hello", "Hello", "Hello")
	enData, err := enPo.MarshalText()
	if err != nil {
		t.Fatalf("failed to marshal english.po: %v", err)
	}
	if err := os.WriteFile(filepath.Join(poDir, "english.po"), enData, 0o644); err != nil {
		t.Fatalf("failed to write english.po: %v", err)
	}

	// Capture output
	var output strings.Builder
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &StatsCmd{
		Input:   csvPath,
		PoDir:   poDir,
		Format:  "text",
		Verbose: false,
	}

	errChan := make(chan error)
	go func() {
		err := cmd.Execute(nil)
		w.Close()
		errChan <- err
	}()

	// Read output
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			output.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	os.Stdout = oldStdout
	if err := <-errChan; err != nil {
		t.Fatalf("StatsCmd.Execute failed: %v", err)
	}

	outputStr := output.String()

	// Verify table header
	if !strings.Contains(outputStr, "Language") {
		t.Error("missing table header")
	}
	if !strings.Contains(outputStr, "Translated") {
		t.Error("missing Translated column in header")
	}

	// Verify russian stats (2 translated out of 4) - table format
	if !strings.Contains(outputStr, "russian") {
		t.Error("missing russian language in output")
	}
	// Check for russian row with correct stats (tabwriter uses spaces for alignment)
	if !strings.Contains(outputStr, "russian") || !strings.Contains(outputStr, "2") || !strings.Contains(outputStr, "4") {
		t.Errorf("missing or incorrect russian stats in table, output: %s", outputStr)
	}
	// Verify percentage is around 50%
	if !strings.Contains(outputStr, "50.0%") {
		t.Errorf("missing correct percentage for russian, output: %s", outputStr)
	}

	// Verify english stats (4 translated out of 4) - table format
	if !strings.Contains(outputStr, "english") {
		t.Error("missing english language in output")
	}
	// Check for english row with correct stats
	if !strings.Contains(outputStr, "english") || !strings.Contains(outputStr, "4") {
		t.Errorf("missing or incorrect english stats in table, output: %s", outputStr)
	}
	// Verify percentage is 100%
	if !strings.Contains(outputStr, "100.0%") {
		t.Errorf("missing correct percentage for english, output: %s", outputStr)
	}
}

// TestStatsCmd_VerboseOutput verifies verbose output with untranslated strings
func TestStatsCmd_VerboseOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample CSV
	csvContent := `"Language","original","russian"
"STR_Yes","Yes","Да"
"STR_No","No",""
`
	csvPath := filepath.Join(tmpDir, "input.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0o644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	// Create PO directory and file
	poDir := filepath.Join(tmpDir, "po")
	if err := os.Mkdir(poDir, 0o755); err != nil {
		t.Fatalf("failed to create po dir: %v", err)
	}

	ruPo := poutil.NewFile()
	ruPo.Language = "russian"
	ruPo.SetHeader("Language", "russian")
	ruPo.SetC("STR_Yes", "Yes", "Да")
	// STR_No is not translated
	ruData, err := ruPo.MarshalText()
	if err != nil {
		t.Fatalf("failed to marshal russian.po: %v", err)
	}
	if err := os.WriteFile(filepath.Join(poDir, "russian.po"), ruData, 0o644); err != nil {
		t.Fatalf("failed to write russian.po: %v", err)
	}

	// Capture output
	var output strings.Builder
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &StatsCmd{
		Input:   csvPath,
		PoDir:   poDir,
		Format:  "text",
		Verbose: true,
	}

	errChan := make(chan error)
	go func() {
		err := cmd.Execute(nil)
		w.Close()
		errChan <- err
	}()

	// Read output
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			output.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	os.Stdout = oldStdout
	if err := <-errChan; err != nil {
		t.Fatalf("StatsCmd.Execute failed: %v", err)
	}

	outputStr := output.String()

	// Verify verbose output contains untranslated strings in grep -nr format
	// Format: po_file:line:key:"original_text"
	if !strings.Contains(outputStr, ".po:") {
		t.Error("missing PO filename in verbose output")
	}
	if !strings.Contains(outputStr, ":STR_No:") {
		t.Errorf("missing untranslated key in output: %s", outputStr)
	}
	if !strings.Contains(outputStr, `:"No"`) {
		t.Errorf("missing untranslated original text in output: %s", outputStr)
	}
}

// TestStatsCmd_JSONOutput verifies JSON format output
func TestStatsCmd_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample CSV
	csvContent := `"Language","original","russian"
"STR_Yes","Yes","Да"
"STR_No","No",""
`
	csvPath := filepath.Join(tmpDir, "input.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0o644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	// Create PO directory and file
	poDir := filepath.Join(tmpDir, "po")
	if err := os.Mkdir(poDir, 0o755); err != nil {
		t.Fatalf("failed to create po dir: %v", err)
	}

	ruPo := poutil.NewFile()
	ruPo.Language = "russian"
	ruPo.SetHeader("Language", "russian")
	ruPo.SetC("STR_Yes", "Yes", "Да")
	ruData, err := ruPo.MarshalText()
	if err != nil {
		t.Fatalf("failed to marshal russian.po: %v", err)
	}
	if err := os.WriteFile(filepath.Join(poDir, "russian.po"), ruData, 0o644); err != nil {
		t.Fatalf("failed to write russian.po: %v", err)
	}

	// Capture output
	var output strings.Builder
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &StatsCmd{
		Input:   csvPath,
		PoDir:   poDir,
		Format:  "json",
		Verbose: true,
	}

	errChan := make(chan error)
	go func() {
		err := cmd.Execute(nil)
		w.Close()
		errChan <- err
	}()

	// Read output
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			output.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	os.Stdout = oldStdout
	if err := <-errChan; err != nil {
		t.Fatalf("StatsCmd.Execute failed: %v", err)
	}

	outputStr := output.String()

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(outputStr), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, outputStr)
	}

	// Verify JSON structure
	languages, ok := result["languages"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing 'languages' key in JSON output: %s", outputStr)
	}

	russian, ok := languages["russian"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing 'russian' language in JSON output: %s", outputStr)
	}

	// Verify stats
	if russian["translated"].(float64) != 1 {
		t.Errorf("expected translated=1, got %v", russian["translated"])
	}
	if russian["total"].(float64) != 2 {
		t.Errorf("expected total=2, got %v", russian["total"])
	}
	if russian["remaining"].(float64) != 1 {
		t.Errorf("expected remaining=1, got %v", russian["remaining"])
	}

	// Verify untranslated array exists in verbose mode
	untranslated, ok := russian["untranslated"].([]interface{})
	if !ok {
		t.Fatalf("missing 'untranslated' array in verbose JSON output")
	}
	if len(untranslated) != 1 {
		t.Errorf("expected 1 untranslated item, got %d", len(untranslated))
	}
}

// TestStatsCmd_LangFilter verifies language filtering
func TestStatsCmd_LangFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample CSV
	csvContent := `"Language","original","russian","english"
"STR_Yes","Yes","Да","Yes"
`
	csvPath := filepath.Join(tmpDir, "input.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0o644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	// Create PO directory and files
	poDir := filepath.Join(tmpDir, "po")
	if err := os.Mkdir(poDir, 0o755); err != nil {
		t.Fatalf("failed to create po dir: %v", err)
	}

	ruPo := poutil.NewFile()
	ruPo.Language = "russian"
	ruPo.SetHeader("Language", "russian")
	ruPo.SetC("STR_Yes", "Yes", "Да")
	ruData, err := ruPo.MarshalText()
	if err != nil {
		t.Fatalf("failed to marshal russian.po: %v", err)
	}
	if err := os.WriteFile(filepath.Join(poDir, "russian.po"), ruData, 0o644); err != nil {
		t.Fatalf("failed to write russian.po: %v", err)
	}

	enPo := poutil.NewFile()
	enPo.Language = "english"
	enPo.SetHeader("Language", "english")
	enPo.SetC("STR_Yes", "Yes", "Yes")
	enData, err := enPo.MarshalText()
	if err != nil {
		t.Fatalf("failed to marshal english.po: %v", err)
	}
	if err := os.WriteFile(filepath.Join(poDir, "english.po"), enData, 0o644); err != nil {
		t.Fatalf("failed to write english.po: %v", err)
	}

	// Capture output
	var output strings.Builder
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &StatsCmd{
		Input:   csvPath,
		PoDir:   poDir,
		Langs:   []string{"russian"},
		Format:  "text",
		Verbose: false,
	}

	errChan := make(chan error)
	go func() {
		err := cmd.Execute(nil)
		w.Close()
		errChan <- err
	}()

	// Read output
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			output.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	os.Stdout = oldStdout
	if err := <-errChan; err != nil {
		t.Fatalf("StatsCmd.Execute failed: %v", err)
	}

	outputStr := output.String()

	// Verify table header
	if !strings.Contains(outputStr, "Language") {
		t.Error("missing table header")
	}

	// Verify only russian is shown - table format
	if !strings.Contains(outputStr, "russian") {
		t.Error("missing russian language in filtered output")
	}
	if strings.Contains(outputStr, "english") {
		t.Error("english should not appear in filtered output")
	}
}

// TestStatsCmd_NoPOFiles verifies error handling when no PO files exist
func TestStatsCmd_NoPOFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample CSV
	csvContent := `"Language","original"
"STR_Yes","Yes"
`
	csvPath := filepath.Join(tmpDir, "input.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0o644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	// Create empty PO directory
	poDir := filepath.Join(tmpDir, "po")
	if err := os.Mkdir(poDir, 0o755); err != nil {
		t.Fatalf("failed to create po dir: %v", err)
	}

	cmd := &StatsCmd{
		Input: csvPath,
		PoDir: poDir,
	}

	err := cmd.Execute(nil)
	if err == nil {
		t.Error("expected error when no PO files exist")
	}
	if !strings.Contains(err.Error(), "no PO files found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestStatsCmd_InvalidLang verifies error handling for invalid language filter
func TestStatsCmd_InvalidLang(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample CSV
	csvContent := `"Language","original"
"STR_Yes","Yes"
`
	csvPath := filepath.Join(tmpDir, "input.csv")
	if err := os.WriteFile(csvPath, []byte(csvContent), 0o644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	// Create PO directory with different language
	poDir := filepath.Join(tmpDir, "po")
	if err := os.Mkdir(poDir, 0o755); err != nil {
		t.Fatalf("failed to create po dir: %v", err)
	}

	ruPo := poutil.NewFile()
	ruPo.Language = "russian"
	ruPo.SetHeader("Language", "russian")
	ruData, err := ruPo.MarshalText()
	if err != nil {
		t.Fatalf("failed to marshal russian.po: %v", err)
	}
	if err := os.WriteFile(filepath.Join(poDir, "russian.po"), ruData, 0o644); err != nil {
		t.Fatalf("failed to write russian.po: %v", err)
	}

	cmd := &StatsCmd{
		Input: csvPath,
		PoDir: poDir,
		Langs: []string{"nonexistent"},
	}

	err = cmd.Execute(nil)
	if err == nil {
		t.Error("expected error when filtering by nonexistent language")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}
