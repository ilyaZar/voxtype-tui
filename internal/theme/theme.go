package theme

import (
	"os"
	"regexp"
)

type Colors struct {
	Accent              string
	Foreground          string
	Background          string
	SelectionForeground string
	SelectionBackground string
	Yellow              string
}

func Load(path string) Colors {
	colors := Colors{
		Accent:              "#81a1c1",
		Foreground:          "#d8dee9",
		Background:          "#2e3440",
		SelectionForeground: "#d8dee9",
		SelectionBackground: "#4c566a",
		Yellow:              "#ebcb8b",
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return colors
	}
	text := string(data)
	colors.Accent = colorValue(text, "accent", colors.Accent)
	colors.Foreground = colorValue(text, "foreground", colors.Foreground)
	colors.Background = colorValue(text, "background", colors.Background)
	colors.SelectionForeground = colorValue(text, "selection_foreground", colors.SelectionForeground)
	colors.SelectionBackground = colorValue(text, "selection_background", colors.SelectionBackground)
	colors.Yellow = colorValue(text, "color3", colors.Yellow)
	return colors
}

func colorValue(text string, key string, fallback string) string {
	re := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(key) + `\s*=\s*"(#[0-9a-fA-F]{6})"`)
	match := re.FindStringSubmatch(text)
	if len(match) != 2 {
		return fallback
	}
	return match[1]
}
