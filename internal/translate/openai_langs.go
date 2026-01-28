package translate

import "strings"

// OpenAITargetLang maps DayZ language names to prompt-friendly language names.
// It returns the original value when the language is not in the map.
func OpenAITargetLang(lang string) string {
	if name, ok := openAILangMap[strings.ToLower(lang)]; ok {
		return name
	}

	return lang
}

var openAILangMap = map[string]string{
	"english":     "English",
	"czech":       "Czech",
	"german":      "German",
	"russian":     "Russian",
	"polish":      "Polish",
	"hungarian":   "Hungarian",
	"italian":     "Italian",
	"spanish":     "Spanish",
	"french":      "French",
	"chinese":     "Chinese",
	"chinesesimp": "Simplified Chinese",
	"japanese":    "Japanese",
	"portuguese":  "Portuguese",
}
