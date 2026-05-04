package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Paths struct {
	ConfigFile string
	CycleFile  string
	ThemeFile  string
	BaseModel  string
}

type AppConfig struct {
	Voxtype VoxtypeConfig `toml:"voxtype"`
	Theme   ThemeConfig   `toml:"theme"`
	Popup   PopupConfig   `toml:"popup"`
}

type VoxtypeConfig struct {
	Config    string `toml:"config"`
	Cycle     string `toml:"cycle"`
	BaseModel string `toml:"base_model"`
	Command   string `toml:"command"`
}

type ThemeConfig struct {
	Colors string `toml:"colors"`
}

type PopupConfig struct {
	Terminal             string `toml:"terminal"`
	Class                string `toml:"class"`
	Title                string `toml:"title"`
	StageWorkspace       string `toml:"stage_workspace"`
	Width                int    `toml:"width"`
	Height               int    `toml:"height"`
	MinWidth             int    `toml:"min_width"`
	MinHeight            int    `toml:"min_height"`
	MarginX              int    `toml:"margin_x"`
	MarginY              int    `toml:"margin_y"`
	TerminalPaddingX     int    `toml:"terminal_padding_x"`
	TerminalPaddingY     int    `toml:"terminal_padding_y"`
	TerminalColumns      int    `toml:"terminal_columns"`
	TerminalRows         int    `toml:"terminal_rows"`
	WaitAttempts         int    `toml:"wait_attempts"`
	WaitIntervalMS       int    `toml:"wait_interval_ms"`
	PositionAttempts     int    `toml:"position_attempts"`
	PositionSettleWaitMS int    `toml:"position_settle_wait_ms"`
	Opacity              string `toml:"opacity"`
}

func DefaultPaths() Paths {
	appConfig := DefaultAppConfig()
	return appConfig.Paths()
}

func DefaultConfigFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	return filepath.Join(home, ".config/voxtype-tui/config.toml")
}

func DefaultAppConfig() AppConfig {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	return AppConfig{
		Voxtype: VoxtypeConfig{
			Config:    filepath.Join(home, ".config/voxtype/config.toml"),
			Cycle:     filepath.Join(home, ".config/voxtype/language_cycle.toml"),
			BaseModel: filepath.Join(home, ".local/share/voxtype/models/ggml-base.bin"),
			Command:   "voxtype",
		},
		Theme: ThemeConfig{
			Colors: filepath.Join(home, ".config/omarchy/current/theme/colors.toml"),
		},
		Popup: PopupConfig{
			Terminal:             "ghostty",
			Class:                "org.omarchy.voxtype.lang-menu",
			Title:                "omarchy-voxtype-language-menu",
			StageWorkspace:       "special:voxtype-tui",
			Width:                500,
			Height:               126,
			MinWidth:             1,
			MinHeight:            1,
			MarginX:              8,
			MarginY:              8,
			TerminalPaddingX:     6,
			TerminalPaddingY:     6,
			TerminalColumns:      46,
			TerminalRows:         7,
			WaitAttempts:         100,
			WaitIntervalMS:       50,
			PositionAttempts:     3,
			PositionSettleWaitMS: 80,
			Opacity:              "0.97 0.94",
		},
	}
}

func LoadAppConfig(path string) (AppConfig, error) {
	if path == "" {
		path = getenv("VOXTYPE_TUI_CONFIG", DefaultConfigFile())
	}

	cfg := DefaultAppConfig()
	data, err := os.ReadFile(expandPath(path))
	if os.IsNotExist(err) {
		applyEnvOverrides(&cfg)
		return cfg, nil
	}
	if err != nil {
		return AppConfig{}, fmt.Errorf("read config: %w", err)
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}, fmt.Errorf("parse config: %w", err)
	}

	cfg.expandPaths()
	applyEnvOverrides(&cfg)
	return cfg, nil
}

func (c AppConfig) Paths() Paths {
	return Paths{
		ConfigFile: c.Voxtype.Config,
		CycleFile:  c.Voxtype.Cycle,
		ThemeFile:  c.Theme.Colors,
		BaseModel:  c.Voxtype.BaseModel,
	}
}

func (c *AppConfig) expandPaths() {
	c.Voxtype.Config = expandPath(c.Voxtype.Config)
	c.Voxtype.Cycle = expandPath(c.Voxtype.Cycle)
	c.Voxtype.BaseModel = expandPath(c.Voxtype.BaseModel)
	c.Theme.Colors = expandPath(c.Theme.Colors)
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

func applyEnvOverrides(c *AppConfig) {
	c.Voxtype.Config = getenv("VOXTYPE_CONFIG", c.Voxtype.Config)
	c.Voxtype.Cycle = getenv("VOXTYPE_CYCLE", c.Voxtype.Cycle)
	c.Voxtype.BaseModel = getenv("VOXTYPE_BASE_MODEL", c.Voxtype.BaseModel)
	c.Theme.Colors = getenv("VOXTYPE_THEME_COLORS", c.Theme.Colors)
	c.expandPaths()
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
