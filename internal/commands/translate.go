package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/woozymasta/dayz-stringtable/internal/poutil"
	"github.com/woozymasta/dayz-stringtable/internal/translate"
)

// TranslateCmd groups subcommands for machine translation providers.
//
// Usage: dayz-stringtable translate [--podir l18n] [--lang russian] [--batch 25] <provider> [provider options]
type TranslateCmd struct {
	Deepl   TranslateDeeplCmd  `command:"deepl" description:"Translate using DeepL"`
	OpenAI  TranslateOpenAICmd `command:"openai" description:"Translate using OpenAI-compatible API"`
	Google  TranslateGoogleCmd `command:"google" description:"Translate using Google Translate"`
	PoDir   string             `short:"d" long:"podir" description:"Directory for PO files" default:"l18n"`
	Langs   []string           `short:"l" long:"lang" description:"Filter by languages (repeatable)"`
	Exclude []string           `short:"e" long:"exclude-lang" description:"Exclude languages (repeatable)"`
	Batch   int                `short:"b" long:"batch" description:"Strings per request batch" default:"25"`
	DryRun  bool               `short:"D" long:"dry-run" description:"Show what would be translated without calling providers"`
}

// NewTranslateCmd wires shared config into subcommands.
func NewTranslateCmd() *TranslateCmd {
	cmd := &TranslateCmd{}
	cmd.Deepl.Common = cmd
	cmd.OpenAI.Common = cmd
	cmd.Google.Common = cmd
	return cmd
}

// Execute runs when translate is called without a provider subcommand.
func (cmd *TranslateCmd) Execute(_ []string) error {
	return fmt.Errorf("missing provider: use 'translate deepl', 'translate openai', or 'translate google'")
}

// TranslateDeeplCmd translates PO files using DeepL.
type TranslateDeeplCmd struct {
	Common *TranslateCmd `no-flag:"true"`

	AuthKey            string `long:"auth-key" description:"DeepL API auth key" env:"DEEPL_AUTH_KEY" default-mask:"-"`
	URL                string `long:"url" description:"DeepL API URL (overrides api-free/api)" env:"DEEPL_API_URL"`
	SourceLang         string `long:"source" description:"Override source language code (e.g. EN)"`
	Formality          string `long:"formality" description:"Tone, if supported by target language" choice:"default" choice:"less" choice:"more"`
	SplitSentences     string `long:"split-sentences" description:"Sentence splitting" choice:"0" choice:"1" choice:"nonewlines"`
	FreeAPI            bool   `long:"api-free" description:"Use api-free.deepl.com endpoint"`
	PreserveFormatting bool   `long:"preserve-formatting" description:"Preserve formatting"`
}

// Execute runs DeepL translation for selected languages.
func (cmd *TranslateDeeplCmd) Execute(_ []string) error {
	common, err := requireCommon(cmd.Common)
	if err != nil {
		return err
	}

	client := &translate.DeeplClient{
		URL:                cmd.URL,
		AuthKey:            cmd.AuthKey,
		SourceLang:         cmd.SourceLang,
		Formality:          cmd.Formality,
		PreserveFormatting: cmd.PreserveFormatting,
		SplitSentences:     cmd.SplitSentences,
		UseFreeAPI:         cmd.FreeAPI,
	}

	return runTranslate(common, client, translate.DeeplTargetLang)
}

// TranslateOpenAICmd translates PO files using OpenAI-compatible API.
type TranslateOpenAICmd struct {
	Common *TranslateCmd `no-flag:"true"`

	APIKey      string  `long:"api-key" description:"OpenAI-compatible API key" env:"OPENAI_API_KEY" default-mask:"-"`
	BaseURL     string  `long:"url" description:"OpenAI-compatible base URL" env:"OPENAI_BASE_URL"`
	Model       string  `long:"model" description:"Model name" default:"gpt-4o-mini"`
	SourceLang  string  `long:"source" description:"Override source language (for prompt)"`
	Temperature float64 `long:"temperature" description:"Sampling temperature"`
}

// Execute runs OpenAI-compatible translation for selected languages.
func (cmd *TranslateOpenAICmd) Execute(_ []string) error {
	common, err := requireCommon(cmd.Common)
	if err != nil {
		return err
	}

	client := &translate.OpenAIClient{
		BaseURL:     cmd.BaseURL,
		APIKey:      cmd.APIKey,
		Model:       cmd.Model,
		Temperature: cmd.Temperature,
	}

	resolve := func(lang string) (string, error) {
		return translate.OpenAITargetLang(lang), nil
	}
	return runTranslateWithSource(common, client, resolve, cmd.SourceLang)
}

// TranslateGoogleCmd translates PO files using Google Translate.
type TranslateGoogleCmd struct {
	Common *TranslateCmd `no-flag:"true"`

	APIKey     string `long:"api-key" description:"Google Translate API key" env:"GOOGLE_TRANSLATE_API_KEY" default-mask:"-"`
	URL        string `long:"url" description:"Google Translate API URL" env:"GOOGLE_TRANSLATE_API_URL"`
	SourceLang string `long:"source" description:"Override source language code (e.g. en)"`
	Format     string `long:"format" description:"Text format" choice:"text" choice:"html" default:"text"`
}

// Execute runs Google Translate for selected languages.
func (cmd *TranslateGoogleCmd) Execute(_ []string) error {
	common, err := requireCommon(cmd.Common)
	if err != nil {
		return err
	}

	client := &translate.GoogleClient{
		URL:        cmd.URL,
		APIKey:     cmd.APIKey,
		SourceLang: cmd.SourceLang,
		Format:     cmd.Format,
	}

	resolve := func(lang string) (string, error) {
		return translate.GoogleTargetLang(lang), nil
	}

	return runTranslateWithSource(common, client, resolve, cmd.SourceLang)
}

// requireCommon validates shared translate flags for provider subcommands.
func requireCommon(common *TranslateCmd) (*TranslateCmd, error) {
	if common == nil {
		return nil, fmt.Errorf("translate command not initialized")
	}
	if common.Batch <= 0 {
		return nil, fmt.Errorf("batch size must be > 0")
	}
	return common, nil
}

// runTranslate executes translation with a provider-specific target resolver.
func runTranslate(common *TranslateCmd, client translate.Client, resolveTarget func(string) (string, error)) error {
	return runTranslateWithSource(common, client, resolveTarget, "")
}

// runTranslateWithSource handles the per-language loop and optional dry-run.
func runTranslateWithSource(common *TranslateCmd, client translate.Client, resolveTarget func(string) (string, error), sourceLang string) error {
	poFiles, err := listPOFiles(common.PoDir)
	if err != nil {
		return err
	}
	if len(poFiles) == 0 {
		return fmt.Errorf("no PO files found in %s", common.PoDir)
	}

	langs, err := selectLangs(common.Langs, common.Exclude, poFiles)
	if err != nil {
		return err
	}
	if len(langs) == 0 {
		return fmt.Errorf("no languages selected after filters")
	}

	ctx := context.Background()
	total := 0
	for _, lang := range langs {
		path := poFiles[lang]
		po, err := poutil.ParseFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		target, err := resolveTarget(lang)
		if err != nil {
			return err
		}

		if common.DryRun {
			count, chars := countPending(po)
			if count == 0 {
				fmt.Printf("lang %s -> %s: nothing to translate\n", lang, target)
				continue
			}
			fmt.Printf("lang %s -> %s: strings %d, chars %d\n", lang, target, count, chars)
			total += count
			continue
		}

		translated, err := translatePO(ctx, po, client, sourceLang, target, common.Batch)
		if err != nil {
			return fmt.Errorf("translate %s: %w", lang, err)
		}
		if translated > 0 {
			po.UpdateBuildHeaders("")
			if err := writePO(path, po); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
		}
		fmt.Printf("lang %s: translated %d\n", lang, translated)
		total += translated
	}

	if total == 0 {
		fmt.Println("no untranslated entries found")
	}
	return nil
}

// translatePO batches untranslated msgid values and writes msgstr responses.
func translatePO(ctx context.Context, po *poutil.File, client translate.Client, sourceLang, targetLang string, batch int) (int, error) {
	var pending []*poutil.Entry
	for _, entry := range po.Entries {
		if entry.MsgStr != "" || entry.MsgID == "" {
			continue
		}
		if entry.HasNoTranslate() {
			continue
		}
		pending = append(pending, entry)
	}
	if len(pending) == 0 {
		return 0, nil
	}

	translated := 0
	for i := 0; i < len(pending); i += batch {
		end := i + batch
		if end > len(pending) {
			end = len(pending)
		}
		texts := make([]string, 0, end-i)
		for _, entry := range pending[i:end] {
			texts = append(texts, entry.MsgID)
		}
		out, err := client.Translate(ctx, translate.Request{
			SourceLang: sourceLang,
			TargetLang: targetLang,
			Texts:      texts,
		})
		if err != nil {
			return translated, err
		}
		if len(out) != len(texts) {
			return translated, fmt.Errorf("translation response size mismatch: got %d, want %d", len(out), len(texts))
		}
		for idx, translatedText := range out {
			pending[i+idx].MsgStr = translatedText
			translated++
		}
	}
	return translated, nil
}

// countPending returns the number of untranslated entries and total rune count.
func countPending(po *poutil.File) (int, int) {
	count := 0
	chars := 0
	for _, entry := range po.Entries {
		if entry.MsgStr != "" || entry.MsgID == "" {
			continue
		}
		if entry.HasNoTranslate() {
			continue
		}
		count++
		chars += len([]rune(entry.MsgID))
	}

	return count, chars
}

// writePO persists the updated PO file in place.
func writePO(path string, po *poutil.File) error {
	data, err := po.MarshalText()
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// listPOFiles maps language names to PO file paths.
func listPOFiles(dir string) (map[string]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.po"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob PO files: %w", err)
	}
	poFiles := make(map[string]string, len(files))
	for _, path := range files {
		lang := ExtractLanguageName(path)
		poFiles[lang] = path
	}
	return poFiles, nil
}

// selectLangs resolves the final language list with includes/excludes applied.
func selectLangs(filter []string, exclude []string, poFiles map[string]string) ([]string, error) {
	normalized := normalizeLangs(filter)
	excluded := normalizeLangs(exclude)
	excludeSet := make(map[string]bool, len(excluded))
	for _, item := range excluded {
		excludeSet[item] = true
	}
	if len(normalized) == 0 {
		return filterExcluded(selectLangsInOrder(poFiles, nil), excludeSet), nil
	}
	var langs []string
	for _, lang := range normalized {
		if _, ok := poFiles[lang]; !ok {
			return nil, fmt.Errorf("language '%s' not found in PO directory", lang)
		}
		langs = append(langs, lang)
	}
	return filterExcluded(selectLangsInOrder(poFiles, langs), excludeSet), nil
}

// normalizeLangs splits comma-separated items and trims whitespace.
func normalizeLangs(items []string) []string {
	var out []string
	for _, item := range items {
		for _, part := range strings.Split(item, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

// selectLangsInOrder keeps the default order while honoring filters.
func selectLangsInOrder(poFiles map[string]string, filter []string) []string {
	if len(filter) > 0 {
		var langs []string
		for _, lang := range filter {
			if _, ok := poFiles[lang]; ok {
				langs = append(langs, lang)
			}
		}
		return langs
	}

	var langs []string
	for _, l := range DefaultLanguages {
		if _, ok := poFiles[l]; ok {
			langs = append(langs, l)
		}
	}
	for l := range poFiles {
		if !ContainsLanguage(DefaultLanguages, l) {
			langs = append(langs, l)
		}
	}
	return langs
}

// filterExcluded removes excluded languages while keeping order.
func filterExcluded(langs []string, excludeSet map[string]bool) []string {
	if len(excludeSet) == 0 {
		return langs
	}
	filtered := make([]string, 0, len(langs))
	for _, lang := range langs {
		if excludeSet[lang] {
			continue
		}
		filtered = append(filtered, lang)
	}

	return filtered
}
