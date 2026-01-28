package translate

import (
	"fmt"
	"strings"
)

// DeeplTargetLang maps DayZ language names to DeepL target codes.
// It returns an error when a language is unsupported by the mapper.
func DeeplTargetLang(lang string) (string, error) {
	if code, ok := deeplLangMap[strings.ToLower(lang)]; ok {
		return code, nil
	}

	return "", fmt.Errorf("unsupported language for deepl: %s", lang)
}

var deeplLangMap = map[string]string{
	"english":     "EN",
	"czech":       "CS",
	"german":      "DE",
	"russian":     "RU",
	"polish":      "PL",
	"hungarian":   "HU",
	"italian":     "IT",
	"spanish":     "ES",
	"french":      "FR",
	"chinese":     "ZH",
	"chinesesimp": "ZH",
	"japanese":    "JA",
	"portuguese":  "PT",
}
