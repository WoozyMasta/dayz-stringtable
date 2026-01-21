package commands

//lint:file-ignore SA5008 go-flags requires duplicate choice tags on struct fields

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/woozymasta/dayz-stringtable/internal/csvutil"
	"github.com/woozymasta/dayz-stringtable/internal/poutil"
)

// StatsCmd displays translation statistics for PO files.
//
// Usage: dayz-stringtable stats --input stringtable.csv --podir l18n [--lang russian] [--verbose] [--format json] [--clear-only]
type StatsCmd struct {
	Input     string   `short:"i" long:"input" description:"CSV input file" default:"stringtable.csv"`
	PoDir     string   `short:"d" long:"podir" description:"Directory for PO files" default:"l18n"`
	Format    string   `short:"f" long:"format" description:"Output format" default:"text" choice:"text" choice:"json"`
	Langs     []string `short:"l" long:"lang" description:"Filter by specific language (all if empty)"`
	Verbose   bool     `short:"V" long:"verbose" description:"Show detailed untranslated strings"`
	ClearOnly bool     `short:"c" long:"clear-only" description:"Don't add notranslate comment, just clear msgstr"`
}

// LangStats holds translation statistics for a single language.
type LangStats struct {
	Language     string             // Language code (e.g., "russian", "english")
	Untranslated []UntranslatedItem // List of untranslated items (only in verbose mode)
	Translated   int                // Number of translated strings
	Total        int                // Total number of strings
	Percentage   float64            // Translation completion percentage
	Remaining    int                // Number of untranslated strings
}

// UntranslatedItem represents a single untranslated string entry.
type UntranslatedItem struct {
	Key      string `json:"key"`      // Translation key (msgid in PO)
	Original string `json:"original"` // Original text (msgctxt in PO)
	Context  string `json:"context"`  // Context (same as original in this implementation)
	PoFile   string `json:"po_file"`  // PO file name (e.g., "russian.po")
	Row      int    `json:"row"`      // CSV row number (1-based, including header)
	PoLine   int    `json:"po_line"`  // Line number in PO file where msgctxt is located
}

// Execute reads CSV and PO files, then displays translation statistics.
func (cmd *StatsCmd) Execute(_ []string) error {
	if cmd.Format == "" {
		cmd.Format = "text"
	}

	rows, err := csvutil.LoadCSV(cmd.Input)
	if err != nil {
		return fmt.Errorf("failed to load CSV: %w", err)
	}

	if len(rows) < 2 {
		return fmt.Errorf("CSV must have header and at least one data row")
	}

	poMap, err := poutil.LoadPODirectory(cmd.PoDir)
	if err != nil {
		return fmt.Errorf("failed to load PO files: %w", err)
	}

	// Build map of language to file path for verbose output
	poFileMap := make(map[string]string)
	files, err := filepath.Glob(filepath.Join(cmd.PoDir, "*.po"))
	if err != nil {
		return fmt.Errorf("failed to glob PO files: %w", err)
	}
	for _, f := range files {
		lang := ExtractLanguageName(f)
		if _, ok := poMap[lang]; ok {
			poFileMap[lang] = f
		}
	}

	langs, err := cmd.selectLanguages(poMap)
	if err != nil {
		return err
	}
	if len(langs) == 0 {
		return fmt.Errorf("no PO files found in directory '%s'", cmd.PoDir)
	}

	allStats := cmd.calculateStats(rows, langs, poMap, poFileMap)

	if cmd.Format == "json" {
		return cmd.outputJSON(allStats)
	}
	return cmd.outputText(allStats)
}

// selectLanguages returns the list of languages to process based on the filter.
// Returns an error if a specific language is requested but not found.
func (cmd *StatsCmd) selectLanguages(poMap map[string]*poutil.File) ([]string, error) {
	if len(cmd.Langs) > 0 {
		var langs []string
		for _, lang := range cmd.Langs {
			if _, ok := poMap[lang]; !ok {
				return nil, fmt.Errorf("language '%s' not found in PO directory", lang)
			}
			langs = append(langs, lang)
		}
		return langs, nil
	}

	return selectLanguagesInOrder(poMap, nil), nil
}

// selectLanguagesInOrder returns languages in default order, then any remaining.
func selectLanguagesInOrder(poMap map[string]*poutil.File, filter []string) []string {
	if len(filter) > 0 {
		var langs []string
		for _, lang := range filter {
			if _, ok := poMap[lang]; ok {
				langs = append(langs, lang)
			}
		}
		return langs
	}

	var langs []string
	// Process languages in default order first
	for _, l := range DefaultLanguages {
		if _, ok := poMap[l]; ok {
			langs = append(langs, l)
		}
	}
	// Add any remaining languages not in default list
	for l := range poMap {
		if !ContainsLanguage(DefaultLanguages, l) {
			langs = append(langs, l)
		}
	}
	return langs
}

// calculateStats computes translation statistics for all specified languages.
func (cmd *StatsCmd) calculateStats(rows [][]string, langs []string, poMap map[string]*poutil.File, poFileMap map[string]string) map[string]*LangStats {
	allStats := make(map[string]*LangStats)
	totalRows := len(rows) - 1

	for _, lang := range langs {
		po := poMap[lang]
		stats := &LangStats{
			Language:     lang,
			Total:        totalRows,
			Untranslated: []UntranslatedItem{},
		}

		for i, row := range rows[1:] {
			if len(row) < 2 {
				continue
			}

			// CSV structure: row[0] = key, row[1] = original text
			// PO format: msgctxt = key, msgid = original text
			key := row[0]
			original := row[1]

			isTranslated := false
			if po != nil {
				// Check if translated (has non-empty msgstr)
				if po.IsTranslatedC(key, original) {
					isTranslated = true
				} else if !cmd.ClearOnly {
					// If --clear-only is not set, entries with # notranslate comment
					// are considered translated (they were intentionally marked as not needing translation)
					entry := po.GetEntry(key, original)
					if entry != nil && entry.HasNoTranslate() {
						isTranslated = true
					}
				}
			}

			if isTranslated {
				stats.Translated++
			} else {
				stats.Remaining++
				if cmd.Verbose {
					poFile := poFileMap[lang]
					poLine := findMsgctxtLine(poFile, key)
					stats.Untranslated = append(stats.Untranslated, UntranslatedItem{
						Row:      i + 2,
						Key:      key,
						Original: original,
						Context:  key,
						PoFile:   filepath.Base(poFile),
						PoLine:   poLine,
					})
				}
			}
		}

		if stats.Total > 0 {
			stats.Percentage = float64(stats.Translated) / float64(stats.Total) * 100.0
		}

		allStats[lang] = stats
	}

	return allStats
}

// outputText outputs statistics in human-readable text format.
// In verbose mode, it shows only untranslated strings in grep -nr format.
// Otherwise, it displays a formatted table with statistics.
func (cmd *StatsCmd) outputText(allStats map[string]*LangStats) error {
	if cmd.Verbose {
		for _, lang := range getSortedLangs(allStats) {
			stats := allStats[lang]
			for _, item := range stats.Untranslated {
				// Format: po_file:line:key:"original_text" (grep -nr style)
				fmt.Printf("%s:%d:%s:%q\n", item.PoFile, item.PoLine, item.Key, item.Original)
			}
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() {
		_ = w.Flush()
	}()

	if _, err := fmt.Fprintln(w, "Language\tTranslated\tTotal\tPercentage\tRemaining"); err != nil {
		return fmt.Errorf("failed to write table header: %w", err)
	}

	for _, lang := range getSortedLangs(allStats) {
		stats := allStats[lang]
		if _, err := fmt.Fprintf(w, "%s\t%d\t%d\t%.1f%%\t%d\n",
			stats.Language, stats.Translated, stats.Total, stats.Percentage, stats.Remaining); err != nil {
			return fmt.Errorf("failed to write table row: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush table: %w", err)
	}

	return nil
}

// outputJSON outputs statistics in JSON format suitable for automation and AI agents.
func (cmd *StatsCmd) outputJSON(allStats map[string]*LangStats) error {
	result := make(map[string]interface{})
	languages := make(map[string]interface{})

	for _, lang := range getSortedLangs(allStats) {
		stats := allStats[lang]
		langData := map[string]interface{}{
			"translated": stats.Translated,
			"total":      stats.Total,
			"percentage": stats.Percentage,
			"remaining":  stats.Remaining,
		}

		if cmd.Verbose {
			langData["untranslated"] = stats.Untranslated
		}

		languages[lang] = langData
	}

	result["languages"] = languages

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// findMsgctxtLine finds the line number where msgctxt with the given key is located in a PO file.
// Returns 0 if the key is not found or if there's an error reading the file.
func findMsgctxtLine(poFile, key string) int {
	if poFile == "" {
		return 0
	}

	file, err := os.Open(poFile)
	if err != nil {
		return 0
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	lookingForMsgctxt := false

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "msgctxt ") {
			if strings.Contains(line, `"`+key+`"`) {
				return lineNum
			}
			lookingForMsgctxt = true
		} else if lookingForMsgctxt {
			// Handle multi-line msgctxt entries
			if strings.HasPrefix(line, `"`) && strings.Contains(line, key) {
				return lineNum
			}
			// Stop looking if we encounter msgid (next entry)
			if strings.HasPrefix(line, "msgid ") {
				lookingForMsgctxt = false
			}
		}
	}

	return 0
}

// getSortedLangs returns languages in a consistent order: first in default order, then any remaining.
func getSortedLangs(allStats map[string]*LangStats) []string {
	var langs []string
	for _, dl := range DefaultLanguages {
		if _, ok := allStats[dl]; ok {
			langs = append(langs, dl)
		}
	}
	for lang := range allStats {
		if !ContainsLanguage(DefaultLanguages, lang) {
			langs = append(langs, lang)
		}
	}
	return langs
}
