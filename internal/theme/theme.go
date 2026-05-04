package theme

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Colors struct {
	Accent              string `toml:"accent"`
	Foreground          string `toml:"foreground"`
	Background          string `toml:"background"`
	SelectionForeground string `toml:"selection_foreground"`
	SelectionBackground string `toml:"selection_background"`
	Red                 string `toml:"color1"`
	Yellow              string `toml:"color3"`
}

func Load(path string) Colors {
	colors := defaultColors()

	for _, candidate := range colorFiles(path) {
		data, err := os.ReadFile(expandPath(candidate))
		if err != nil {
			continue
		}
		if err := toml.Unmarshal(data, &colors); err != nil {
			return defaultColors()
		}
		return colors
	}
	return colors
}

func defaultColors() Colors {
	return Colors{
		Accent:              "#81a1c1",
		Foreground:          "#d8dee9",
		Background:          "#2e3440",
		SelectionForeground: "#d8dee9",
		SelectionBackground: "#4c566a",
		Red:                 "#bf616a",
		Yellow:              "#ebcb8b",
	}
}

func colorFiles(path string) []string {
	current := defaultOmarchyColorsFile()
	if path == "" || path == current {
		return []string{current}
	}
	return []string{path, current}
}

func defaultOmarchyColorsFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	return filepath.Join(home, ".config/omarchy/current/theme/colors.toml")
}

func expandPath(path string) string {
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	return path
}
