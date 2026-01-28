package translate

import "strings"

// GoogleTargetLang maps DayZ language names to Google Translate target codes.
// It returns the original value when the language is not in the map.
func GoogleTargetLang(lang string) string {
	if code, ok := googleLangMap[strings.ToLower(lang)]; ok {
		return code
	}

	return lang
}

// googleLangMap holds DayZ language names mapped to Google codes.
var googleLangMap = map[string]string{
	"english":     "en",
	"czech":       "cs",
	"german":      "de",
	"russian":     "ru",
	"polish":      "pl",
	"hungarian":   "hu",
	"italian":     "it",
	"spanish":     "es",
	"french":      "fr",
	"chinese":     "zh",
	"chinesesimp": "zh-CN",
	"japanese":    "ja",
	"portuguese":  "pt",
}
