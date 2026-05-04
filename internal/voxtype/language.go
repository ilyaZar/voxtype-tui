package voxtype

type Language struct {
	Code     string
	Label    string
	Menu     string
	Model    string
	Language string
}

var Languages = []Language{
	{Code: "en", Label: "EN", Menu: "English (EN)", Model: "base.en", Language: "en"},
	{Code: "de", Label: "DE", Menu: "German (DE)", Model: "base", Language: "de"},
	{Code: "ru", Label: "RU", Menu: "Russian (RU)", Model: "base", Language: "ru"},
}

func AllCodes() []string {
	codes := make([]string, 0, len(Languages))
	for _, language := range Languages {
		codes = append(codes, language.Code)
	}
	return codes
}

func ByCode(code string) (Language, bool) {
	for _, language := range Languages {
		if language.Code == code {
			return language, true
		}
	}
	return Language{}, false
}

func KnownCodes(codes []string) []string {
	requested := make(map[string]bool, len(codes))
	for _, code := range codes {
		requested[code] = true
	}

	known := make([]string, 0, len(codes))
	for _, language := range Languages {
		if requested[language.Code] {
			known = append(known, language.Code)
		}
	}
	return known
}

func Labels(codes []string) string {
	labels := ""
	for index, code := range codes {
		language, ok := ByCode(code)
		if !ok {
			continue
		}
		if index > 0 && labels != "" {
			labels += ", "
		}
		labels += language.Label
	}
	return labels
}
