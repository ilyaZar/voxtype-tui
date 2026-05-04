package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ilyaZar/voxtype-tui/internal/voxtype"
)

type WhisperPreset struct {
	Model    string
	Language string
}

func CurrentPreset(path string) (voxtype.Language, bool, error) {
	lines, err := readLines(path)
	if err != nil {
		return voxtype.Language{}, false, err
	}
	preset, err := currentWhisperPreset(lines)
	if err != nil {
		return voxtype.Language{}, false, err
	}

	for _, language := range voxtype.Languages {
		if language.Model == preset.Model && language.Language == preset.Language {
			return language, true, nil
		}
	}
	return voxtype.Language{}, false, nil
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
	return os.WriteFile(path, []byte(strings.Join(updated, "")), 0o644)
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config not found: %s", path)
	}
	return strings.SplitAfter(string(data), "\n"), nil
}

func currentWhisperPreset(lines []string) (WhisperPreset, error) {
	inWhisper := false
	preset := WhisperPreset{}

	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "[") && strings.HasSuffix(stripped, "]") {
			inWhisper = stripped == "[whisper]"
			continue
		}
		if !inWhisper {
			continue
		}

		if match := regexp.MustCompile(`^\s*model\s*=\s*"([^"]+)"`).FindStringSubmatch(line); len(match) == 2 {
			preset.Model = match[1]
		}
		if match := regexp.MustCompile(`^\s*language\s*=\s*"([^"]+)"`).FindStringSubmatch(line); len(match) == 2 {
			preset.Language = match[1]
		}
	}

	if preset.Model == "" || preset.Language == "" {
		return WhisperPreset{}, fmt.Errorf("could not find [whisper] model/language entries")
	}
	return preset, nil
}

func replaceWhisperPreset(lines []string, target voxtype.Language) ([]string, error) {
	inWhisper := false
	replacedModel := false
	replacedLanguage := false
	updated := make([]string, 0, len(lines))

	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "[") && strings.HasSuffix(stripped, "]") {
			inWhisper = stripped == "[whisper]"
			updated = append(updated, line)
			continue
		}

		if inWhisper && !replacedModel && regexp.MustCompile(`^\s*model\s*=`).MatchString(line) {
			indent := regexp.MustCompile(`^\s*`).FindString(line)
			updated = append(updated, fmt.Sprintf("%smodel = %q\n", indent, target.Model))
			replacedModel = true
			continue
		}

		if inWhisper && !replacedLanguage && regexp.MustCompile(`^\s*language\s*=`).MatchString(line) {
			indent := regexp.MustCompile(`^\s*`).FindString(line)
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
