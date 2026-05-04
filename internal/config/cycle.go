package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/ilyaZar/voxtype-tui/internal/voxtype"
	"github.com/pelletier/go-toml/v2"
)

func ReadCycle(path string, languages ...[]voxtype.Language) ([]string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return voxtype.AllCodes(languages...), nil
	}
	if err != nil {
		return nil, err
	}

	var cfg struct {
		Enabled []string `toml:"enabled"`
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse cycle %s: %w", path, err)
	}

	codes := voxtype.KnownCodes(cfg.Enabled, languages...)
	if len(codes) == 0 {
		return voxtype.AllCodes(languages...), nil
	}
	return codes, nil
}

func WriteCycle(path string, codes []string) error {
	if len(codes) == 0 {
		return fmt.Errorf("select at least one language")
	}
	quoted := make([]string, 0, len(codes))
	for _, code := range codes {
		quoted = append(quoted, fmt.Sprintf("%q", code))
	}
	text := "# Languages included in the Ctrl+F12 Voxtype cycle.\n"
	text += fmt.Sprintf("enabled = [%s]\n", strings.Join(quoted, ", "))
	return atomicWriteFile(path, []byte(text), 0o644)
}
