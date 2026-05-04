package voxtype

import "strings"

type Language struct {
	Code     string `toml:"code"`
	Name     string `toml:"name"`
	Label    string `toml:"label"`
	Model    string `toml:"model"`
	Language string `toml:"language"`
}

func AllCodes(languages ...[]Language) []string {
	list := languageList(languages...)
	codes := make([]string, 0, len(list))
	for _, language := range list {
		codes = append(codes, language.Code)
	}
	return codes
}

func ByCode(code string, languages ...[]Language) (Language, bool) {
	for _, language := range languageList(languages...) {
		if language.Code == code {
			return language, true
		}
	}
	return Language{}, false
}

func KnownCodes(codes []string, languages ...[]Language) []string {
	seen := make(map[string]bool, len(codes))
	known := make([]string, 0, len(codes))
	for _, code := range codes {
		if seen[code] {
			continue
		}
		if _, ok := ByCode(code, languages...); ok {
			known = append(known, code)
			seen[code] = true
		}
	}
	return known
}

func Labels(codes []string, languages ...[]Language) string {
	labels := make([]string, 0, len(codes))
	for _, code := range codes {
		language, ok := ByCode(code, languages...)
		if !ok {
			continue
		}
		labels = append(labels, language.Label)
	}
	return strings.Join(labels, ", ")
}

func languageList(languages ...[]Language) []Language {
	if len(languages) > 0 && len(languages[0]) > 0 {
		return languages[0]
	}
	return nil
}
