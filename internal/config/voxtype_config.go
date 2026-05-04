package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ilyaZar/voxtype-tui/internal/voxtype"
	"github.com/pelletier/go-toml/v2"
)

type whisperPreset struct {
	Model    string `toml:"model"`
	Language string `toml:"language"`
}

type tuiPresetConfig struct {
	Language []voxtype.Language `toml:"language"`
}

var (
	tableHeaderRE   = regexp.MustCompile(`^\s*\[{1,2}[^\]]+\]{1,2}\s*(?:#.*)?$`)
	whisperHeaderRE = regexp.MustCompile(`^\s*\[\s*whisper\s*\]\s*(?:#.*)?$`)
	modelKeyRE      = regexp.MustCompile(`^\s*model\s*=`)
	languageKeyRE   = regexp.MustCompile(`^\s*language\s*=`)
	leadingIndentRE = regexp.MustCompile(`^\s*`)
)

func ReadLanguages(path string) ([]voxtype.Language, error) {
	data, err := readConfig(path)
	if err != nil {
		return nil, err
	}
	var cfg struct {
		VoxtypeTUI tuiPresetConfig `toml:"voxtype_tui"`
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	languages := normalizeLanguages(cfg.VoxtypeTUI.Language)
	if len(languages) == 0 {
		return nil, noConfiguredLanguagesError{path: canonicalConfigPath(path)}
	}
	return languages, nil
}

func CurrentPreset(path string, languages ...[]voxtype.Language) (voxtype.Language, bool, error) {
	data, err := readConfig(path)
	if err != nil {
		return voxtype.Language{}, false, err
	}
	var cfg struct {
		Whisper whisperPreset `toml:"whisper"`
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return voxtype.Language{}, false, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.Whisper.Model == "" || cfg.Whisper.Language == "" {
		return voxtype.Language{}, false, fmt.Errorf("could not find [whisper] model/language entries")
	}

	for _, language := range languageList(languages...) {
		if language.Model == cfg.Whisper.Model && language.Language == cfg.Whisper.Language {
			return language, true, nil
		}
	}
	return voxtype.Language{}, false, nil
}

func normalizeLanguages(languages []voxtype.Language) []voxtype.Language {
	seen := make(map[string]bool, len(languages))
	known := make([]voxtype.Language, 0, len(languages))
	for _, language := range languages {
		language.Code = strings.ToLower(strings.TrimSpace(language.Code))
		language.Name = strings.TrimSpace(language.Name)
		language.Label = strings.ToUpper(strings.TrimSpace(language.Label))
		language.Model = strings.TrimSpace(language.Model)
		language.Language = strings.TrimSpace(language.Language)
		if language.Code == "" || language.Model == "" || language.Language == "" || seen[language.Code] {
			continue
		}
		if language.Name == "" {
			language.Name = strings.ToUpper(language.Code)
		}
		if language.Label == "" {
			language.Label = strings.ToUpper(language.Code)
		}
		known = append(known, language)
		seen[language.Code] = true
	}
	return known
}

func languageList(languages ...[]voxtype.Language) []voxtype.Language {
	if len(languages) > 0 && len(languages[0]) > 0 {
		return languages[0]
	}
	return nil
}

func canonicalConfigPath(path string) string {
	path = expandPath(path)
	if absolute, err := filepath.Abs(path); err == nil {
		return absolute
	}
	return path
}

type noConfiguredLanguagesError struct {
	path string
}

func (e noConfiguredLanguagesError) Error() string {
	return "No configured languages in " + e.path
}

func WritePreset(path string, target voxtype.Language) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	updated, err := replaceWhisperPreset(lines, target)
	if err != nil {
		return err
	}
	return atomicWriteFile(path, []byte(strings.Join(updated, "")), 0o644)
}

func readLines(path string) ([]string, error) {
	data, err := readConfig(path)
	if err != nil {
		return nil, err
	}
	return strings.SplitAfter(string(data), "\n"), nil
}

func readConfig(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found: %s", path)
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	return data, nil
}

func replaceWhisperPreset(lines []string, target voxtype.Language) ([]string, error) {
	inWhisper := false
	replacedModel := false
	replacedLanguage := false
	updated := make([]string, 0, len(lines))

	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if tableHeaderRE.MatchString(stripped) {
			inWhisper = whisperHeaderRE.MatchString(stripped)
			updated = append(updated, line)
			continue
		}

		if inWhisper && !replacedModel && modelKeyRE.MatchString(line) {
			indent := leadingIndentRE.FindString(line)
			updated = append(updated, fmt.Sprintf("%smodel = %q\n", indent, target.Model))
			replacedModel = true
			continue
		}

		if inWhisper && !replacedLanguage && languageKeyRE.MatchString(line) {
			indent := leadingIndentRE.FindString(line)
			updated = append(updated, fmt.Sprintf("%slanguage = %q\n", indent, target.Language))
			replacedLanguage = true
			continue
		}

		updated = append(updated, line)
	}

	if !replacedModel || !replacedLanguage {
		return nil, fmt.Errorf("failed to update [whisper] model/language entries")
	}
	return updated, nil
}
